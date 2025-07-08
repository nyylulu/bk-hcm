/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package detector ...
package detector

import (
	"encoding/json"
	"fmt"
	"strings"

	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/converter"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

const (
	_cvmFilterIp            = "private-ip-address"
	_vpcFilterIp            = "address-ip"
	_vpcFilterKeyExactMatch = "ip-exact-match"
	_vpcFilterValExactMatch = "true"
)

var (
	// data from DB: db_so_cvm, table: tb_cvm_zone
	// TODO: should get from cvm service
	_regionMap = map[string]string{
		"上海":   "ap-shanghai",
		"南京":   "ap-nanjing",
		"天津":   "ap-tianjin",
		"广州":   "ap-guangzhou",
		"清远":   "ap-guangzhou",
		"佛山":   "ap-guangzhou",
		"深圳":   "ap-shenzhen",
		"重庆":   "ap-chongqing",
		"香港":   "ap-hongkong",
		"新加坡":  "ap-singapore",
		"孟买":   "ap-mumbai",
		"圣克拉拉": "na-siliconvalley",
		"苏州":   "ap-shanghai",
		"扬州":   "ap-nanjing",
		"首尔":   "ap-seoul",
		"西安":   "ap-xian-ec",
		"郑州":   "ap-zhengzhou-ec",
		"济南":   "ap-jinan-ec",
		"福州":   "ap-fuzhou-ec",
		"长沙":   "ap-changsha-ec",
		"武汉":   "ap-wuhan-ec",
		"法兰克福": "eu-frankfurt",
		"默费尔登": "eu-frankfurt",
		"东京":   "ap-tokyo",
		"曼谷":   "ap-bangkok",
		"莫斯科":  "eu-moscow",
		"石家庄":  "ap-shijiazhuang-ec",
		"杭州":   "ap-hangzhou-ec",
		"巴里":   "na-toronto",
		"北京":   "ap-beijing",
		"成都":   "ap-chengdu",
		"沈阳":   "ap-shenyang-ec",
		"台北":   "ap-taipei", // no vpc
		"合肥":   "ap-hefei-ec",
		"雅加达":  "ap-jakarta", // no vpc
		"汕尾":   "ap-shenzhen",
		"圣保罗":  "sa-saopaulo",
	}
	_vpcIdMap = map[string]string{
		"上海":   "vpc-2x7lhtse",
		"南京":   "vpc-fb7sybzv",
		"天津":   "vpc-1yoew5gc",
		"广州":   "vpc-03nkx9tv",
		"清远":   "vpc-03nkx9tv",
		"佛山":   "vpc-03nkx9tv",
		"深圳":   "vpc-kwgem8tj",
		"重庆":   "vpc-gelpqsur",
		"香港":   "vpc-b5okec48",
		"新加坡":  "vpc-706wf55j",
		"孟买":   "vpc-59eofud4",
		"圣克拉拉": "vpc-n040n5bl",
		"苏州":   "vpc-2x7lhtse",
		"扬州":   "vpc-fb7sybzv",
		"首尔":   "vpc-99wg8fre",
		"西安":   "vpc-efw4kf6r",
		"郑州":   "vpc-54mjeaf8",
		"济南":   "vpc-kgepmcdd",
		"福州":   "vpc-hdxonj2q",
		"长沙":   "vpc-erdqk82h",
		"武汉":   "vpc-867lsj6w",
		"法兰克福": "vpc-38klpz7z",
		"默费尔登": "vpc-38klpz7z",
		"东京":   "vpc-8iple1iq",
		"曼谷":   "vpc-pdnxzhz8",
		"莫斯科":  "vpc-p62yjqvp",
		"石家庄":  "vpc-6b3vbija",
		"杭州":   "vpc-puhasca0",
		"巴里":   "vpc-drefwt2v",
		"北京":   "vpc-bhb0y6g8",
		"成都":   "vpc-r1wicnlq",
		"沈阳":   "vpc-rea7a2kc",
		"合肥":   "vpc-e0a5jxa7",
		"汕尾":   "vpc-kwgem8tj",
		"圣保罗":  "vpc-0ypt4zc1",
	}
)

type tencentCloudClients struct {
	cvmClient *cvm.Client
	vpcClient *vpc.Client
	clbClient *clb.Client
}

// Backend lb backend
type Backend struct {
	VpcId     string `json:"VpcId"`
	PrivateIp string `json:"PrivateIp"`
}

// DescribeLBListenerRequest gets lb listener request
type DescribeLBListenerRequest struct {
	*tchttp.BaseRequest
	Backends []Backend `json:"Backends"`
}

// NewDescribeLBListenersRequest creates a get lb listener request
func NewDescribeLBListenersRequest() (request *DescribeLBListenerRequest) {
	request = &DescribeLBListenerRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("clb", "2018-03-17", "DescribeLBListeners")
	return
}

// ToJsonString encodes to json string
func (r *DescribeLBListenerRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString decodes from json string
func (r *DescribeLBListenerRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// ClassicLoadBalancers classic load balancer
type ClassicLoadBalancers struct {
	LoadBalancerId string                 `json:"LoadBalancerId"`
	Vip            string                 `json:"Vip"`
	Region         string                 `json:"Region"`
	Listeners      []*clb.ListenerBackend `json:"Listeners"`
}

// DescribeLBListenersResponse get lb listener response
type DescribeLBListenersResponse struct {
	*tchttp.BaseResponse
	Response *struct {
		// 唯一请求 ID，每次请求都会返回。定位问题时需要提供该次请求的 RequestId。
		RequestId     *string                 `json:"RequestId,omitempty" name:"RequestId"`
		LoadBalancers []*ClassicLoadBalancers `json:"LoadBalancers"`
	} `json:"Response"`
}

// NewDescribeLoadBalancersResponse creates a get lb listener response
func NewDescribeLoadBalancersResponse() (request *DescribeLBListenersResponse) {
	request = &DescribeLBListenersResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ToJsonString encodes to json string
func (r *DescribeLBListenersResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString decodes from json string
func (r *DescribeLBListenersResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

func initCloudClients(tcOpt cc.TCloudCli, host *cmdb.Host) (*tencentCloudClients, bool, error) {
	id, key := tcOpt.Credential.ID, tcOpt.Credential.Key
	timeOut := 30
	areaEn, ok := _regionMap[host.BkZoneName]
	if !ok {
		return nil, false, fmt.Errorf("predefine area map has no such zone: %s", host.BkZoneName)
	}

	credCvm := common.NewCredential(id, key)
	cpfCvm := profile.NewClientProfile()
	cpfCvm.HttpProfile.Endpoint = tcOpt.Endpoints.Cvm
	cpfCvm.HttpProfile.ReqTimeout = timeOut
	cvmClient, _ := cvm.NewClient(credCvm, areaEn, cpfCvm)

	credVpc := common.NewCredential(id, key)
	cpfVpc := profile.NewClientProfile()
	cpfVpc.HttpProfile.Endpoint = tcOpt.Endpoints.Vpc
	cpfVpc.HttpProfile.ReqTimeout = timeOut
	vpcClient, _ := vpc.NewClient(credVpc, areaEn, cpfVpc)

	credClb := common.NewCredential(id, key)
	cpfClb := profile.NewClientProfile()
	cpfClb.HttpProfile.Endpoint = tcOpt.Endpoints.Clb
	cpfClb.HttpProfile.ReqTimeout = timeOut
	clbClient, _ := clb.NewClient(credClb, areaEn, cpfClb)

	clients := &tencentCloudClients{
		cvmClient: cvmClient,
		vpcClient: vpcClient,
		clbClient: clbClient,
	}

	return clients, false, nil
}

// checkCvmWorkGroup ...
type checkCvmWorkGroup struct {
	baseWorkGroup
}

// newCheckCvmWorkGroup ...
func newCheckCvmWorkGroup(resultHandler StepResultHandler, workerNum int, cliSet *cliSet) *checkCvmWorkGroup {
	return &checkCvmWorkGroup{
		baseWorkGroup: newBaseWorkGroup(enumor.CheckCvmDetectStep, resultHandler, workerNum, checkCvm, cliSet),
	}
}

func checkCvm(kt *kit.Kit, steps []*StepMeta, resultHandler StepResultHandler, cliSet *cliSet) {
	hostIDs := make([]int64, 0)
	for _, step := range steps {
		hostIDs = append(hostIDs, step.Step.HostID)
	}
	ccOp := NewCmdbOperator(cliSet.cc)
	hosts, err := ccOp.GetHostBaseInfoByID(kt, hostIDs)
	if err != nil {
		logs.Errorf("failed to check cvm, for get host from cc err: %v, host id: %v, rid: %s", err, hostIDs, kt.Rid)
		resultHandler.HandleResult(kt, steps, err, err.Error(), true)
		return
	}
	idHostMap := make(map[int64]cmdb.Host)
	for _, host := range hosts {
		idHostMap[host.BkHostID] = host
	}

	for _, step := range steps {
		host, ok := idHostMap[step.Step.HostID]
		if !ok {
			logs.Errorf("failed to check cvm, can not find host, host id: %d, ip: %s, rid: %s", step.Step.HostID,
				step.Step.IP, kt.Rid)
			err = fmt.Errorf("can not find host, host id: %d, ip: %s", step.Step.HostID, step.Step.IP)
			resultHandler.HandleResult(kt, []*StepMeta{step}, err, err.Error(), false)
			continue
		}

		// cvm check rate limit 50, to avoid tencent cloud sdk internal error
		cvmLimiter.Take()
		exeInfo, retry, err := checkCvmStrategy(kt, &host, cliSet)
		if err != nil {
			logs.Errorf("failed to check cvm, err: %v, ip: %s, rid: %s", err, step.Step.IP, kt.Rid)
		}
		resultHandler.HandleResult(kt, []*StepMeta{step}, err, exeInfo, retry)
	}
}

func checkCvmStrategy(kt *kit.Kit, host *cmdb.Host, cliSet *cliSet) (string, bool, error) {
	exeInfos := make([]string, 0)

	// skip cvm check if host is not TC device
	if !isTcDevice(host) {
		return "", false, nil
	}

	clients, retry, err := initCloudClients(cliSet.tcOpt, host)
	if err != nil {
		logs.Errorf("failed to init tencent cloud client, err: %v, ip: %s, rid: %s", err, host.BkHostInnerIP, kt.Rid)
		return strings.Join(exeInfos, "\n"), retry, err
	}

	// check security group and clb for docker on cvm
	if isDockerVM(host) {
		exeInfo, retry, err := checkDockerStrategy(kt, cliSet, clients, host)
		exeInfos = append(exeInfos, exeInfo)
		if err != nil {
			return strings.Join(exeInfos, "\n"), retry, err
		}
		return strings.Join(exeInfos, "\n"), false, nil
	}

	// check security group and clb for cvm
	exeInfo, retry, err := checkVmStrategy(kt, cliSet, clients, host)
	exeInfos = append(exeInfos, exeInfo)
	if err != nil {
		return strings.Join(exeInfos, "\n"), retry, err
	}

	return strings.Join(exeInfos, "\n"), false, nil
}

func checkDockerStrategy(kt *kit.Kit, cliSet *cliSet, clients *tencentCloudClients, host *cmdb.Host) (string, bool,
	error) {

	exeInfos := make([]string, 0)

	exeInfo, retry, err := checkDockerSecurityGroup(kt, cliSet, clients, host)
	exeInfos = append(exeInfos, exeInfo)
	if err != nil {
		return strings.Join(exeInfos, "\n"), retry, err
	}

	exeInfo, retry, err = checkCLB(kt, clients, host)
	exeInfos = append(exeInfos, exeInfo)
	if err != nil {
		return strings.Join(exeInfos, "\n"), retry, err
	}

	return strings.Join(exeInfos, "\n"), false, nil
}

func checkVmStrategy(kt *kit.Kit, cliSet *cliSet, clients *tencentCloudClients, host *cmdb.Host) (string, bool, error) {
	exeInfos := make([]string, 0)

	exeInfo, retry, err := checkVmSecurityGroup(kt, cliSet, clients, host)
	exeInfos = append(exeInfos, exeInfo)
	if err != nil {
		return strings.Join(exeInfos, "\n"), retry, err
	}

	exeInfo, retry, err = checkCLB(kt, clients, host)
	exeInfos = append(exeInfos, exeInfo)
	if err != nil {
		return strings.Join(exeInfos, "\n"), retry, err
	}

	return strings.Join(exeInfos, "\n"), false, nil
}

func checkDockerSecurityGroup(kt *kit.Kit, cliSet *cliSet, clients *tencentCloudClients, host *cmdb.Host) (string, bool,
	error) {

	request := cvm.NewDescribeInstancesRequest()
	filterIp := _cvmFilterIp
	ip := host.GetUniqIp()
	request.Filters = append(request.Filters, &cvm.Filter{
		Name:   &filterIp,
		Values: []*string{&ip},
	})
	resp, err := clients.cvmClient.DescribeInstances(request)
	if err != nil {
		logs.Errorf("cvm describe instance failed, ip: %s, err: %v, rid: %s", ip, err, kt.Rid)
		return "", true, fmt.Errorf("get instances failed: %s", err)
	}

	respStr := structToStr(resp)
	exeInfo := fmt.Sprintf("cvm response: %s", respStr)

	if len(resp.Response.InstanceSet) == 0 {
		return exeInfo, false, nil
	}

	for _, inst := range resp.Response.InstanceSet {
		for _, sgId := range inst.SecurityGroupIds {
			if retry, err := checkIsDefaultSG(kt, cliSet, clients, converter.PtrToVal(sgId)); err != nil {
				return exeInfo, retry, err
			}
		}
	}

	return exeInfo, false, nil
}

func checkIsDefaultSG(kt *kit.Kit, cliSet *cliSet, clients *tencentCloudClients, sgId string) (retry bool, err error) {
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", constant.GlobalConfigTypeRegionDefaultSecGroup),
			tools.RuleJSONEqual("config_value.security_group_id", sgId),
		),
		Page: &core.BasePage{Count: true},
	}
	list, err := cliSet.hcm.DataService().Global.GlobalConfig.List(kt, listReq)
	if err != nil {
		logs.Errorf("failed to get default security group, err: %v, id: %s, rid: %s", err, sgId, kt.Rid)
		return true, err
	}
	if list.Count > 0 {
		return false, nil
	}

	req := vpc.NewDescribeSecurityGroupsRequest()
	req.Filters = append(req.Filters, &vpc.Filter{
		Name:   common.StringPtr("security-group-id"),
		Values: common.StringPtrs([]string{sgId}),
	})
	resp, err := clients.vpcClient.DescribeSecurityGroups(req)
	if err != nil {
		logs.Errorf("vpc describe security group failed, err: %v, sgID: %s, rid: %s", err, sgId, kt.Rid)
		return true, fmt.Errorf("get sg failed: %v", err)
	}
	if len(resp.Response.SecurityGroupSet) == 0 {
		return false, fmt.Errorf("did not find security group: %s", sgId)
	}

	if *resp.Response.SecurityGroupSet[0].SecurityGroupName != "云梯默认安全组" {
		return false, fmt.Errorf("has non-default security group: %s",
			*resp.Response.SecurityGroupSet[0].SecurityGroupName)
	}
	return false, nil
}

func checkVmSecurityGroup(kt *kit.Kit, cliSet *cliSet, clients *tencentCloudClients, host *cmdb.Host) (string, bool,
	error) {

	request := vpc.NewDescribeNetworkInterfacesRequest()
	// params := "{\"Filters\":[{\"Name\":\"private-ip-address\",\"Values\":[\"X.XXX.XXX.XXX\"]}]}"
	// address-ip - String - （过滤条件）内网IPv4地址，单IP后缀模糊匹配，多IP精确匹配。可以与`ip-exact-match`配合做单IP的精确匹配查询。
	// ip-exact-match - Boolean - （过滤条件）内网IPv4精确匹配查询，存在多值情况，只取第一个。
	request.Filters = []*vpc.Filter{
		{
			Name:   common.StringPtr(_vpcFilterIp),
			Values: common.StringPtrs([]string{host.GetUniqIp()}),
		},
		{
			Name:   common.StringPtr(_vpcFilterKeyExactMatch),
			Values: common.StringPtrs([]string{_vpcFilterValExactMatch}),
		},
	}

	resp, err := clients.vpcClient.DescribeNetworkInterfaces(request)
	if err != nil {
		logs.Errorf("vpc describe network interfaces failed, err: %v, ip: %s, rid: %s", err, host.GetUniqIp(), kt.Rid)
		return "", true, fmt.Errorf("failed to check vpc: %v", err)
	}

	respStr := structToStr(resp)
	exeInfo := fmt.Sprintf("vpc response: %s", respStr)

	if len(resp.Response.NetworkInterfaceSet) == 0 {
		return exeInfo, false, nil
	}

	// 任一network interface存在非云梯默认安全组，无法回收
	for _, netIf := range resp.Response.NetworkInterfaceSet {
		for _, gs := range netIf.GroupSet {
			if retry, err := checkIsDefaultSG(kt, cliSet, clients, *gs); err != nil {
				return exeInfo, retry, err
			}
		}
	}

	return exeInfo, false, nil
}

func checkCLB(kt *kit.Kit, clients *tencentCloudClients, host *cmdb.Host) (string, bool, error) {
	request := NewDescribeLBListenersRequest()
	vpcId, ok := _vpcIdMap[host.BkZoneName]
	if !ok {
		return "", false, fmt.Errorf("predefine vpcid map has no such zone: %s", host.BkZoneName)
	}

	request.Backends = []Backend{
		{
			PrivateIp: host.GetUniqIp(),
			VpcId:     vpcId,
		},
	}
	resp := NewDescribeLoadBalancersResponse()
	if err := clients.clbClient.Send(request, resp); err != nil {
		logs.Errorf("check clb failed, ip: %s, vpcId: %s, err: %v, rid: %s", host.GetUniqIp(), vpcId, err, kt.Rid)
		return "", true, fmt.Errorf("failed to check clb, err: %v", err)
	}

	respStr := structToStr(resp)
	exeInfo := fmt.Sprintf("clb response: %s", respStr)

	if len(resp.Response.LoadBalancers) > 0 {
		lbIDs := make([]string, 0)
		for _, lb := range resp.Response.LoadBalancers {
			lbIDs = append(lbIDs, lb.LoadBalancerId)
		}
		lbInfo := strings.Join(lbIDs, ",")
		return exeInfo, false, fmt.Errorf("has binding clb: %s", lbInfo)
	}

	fmt.Println(resp.ToJsonString())
	return exeInfo, false, nil
}
