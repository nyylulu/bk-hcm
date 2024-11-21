/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package ziyan

import (
	"errors"
	"fmt"

	"hcm/pkg/adaptor/poller"
	"hcm/pkg/adaptor/types"
	typelb "hcm/pkg/adaptor/types/load-balancer"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// CreateZiyanLoadBalancer reference: https://cloud.tencent.com/document/api/214/30692
// 如果创建成功返回对应clb id, 需要检查对应的`SuccessCloudIDs`参数。
func (t *ZiyanAdpt) CreateZiyanLoadBalancer(kt *kit.Kit, opt *typelb.TCloudZiyanCreateClbOption) (
	*poller.BaseDoneResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v", opt.Region, err)
	}

	req := t.formatCreateClbRequest(opt)

	createResp, err := client.CreateLoadBalancerWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("create tencent cloud clb instance failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}
	/*
		NOTE：云上接口`CreateLoadBalancer`返回实例`ID`列表并不代表实例创建成功。`CreateLoadBalancer`接口文档声称可根据
		[DescribeLoadBalancers](https://cloud.tencent.com/document/api/214/30685)接口返回的`LoadBalancerSet`中
		对应实例的`ID`的状态来判断创建是否完成：如果实例状态由“0(创建中)”变为“1(正常运行)”，则为创建成功。
		但是实际上对于创建失败的任务使用`DescribeLoadBalancers`接口无法判断，该情况并不会返回错误，只会静默返回空值。
		因此，用`DescribeLoadBalancers`这个接口难以确定是创建时间过长还是创建失败。
		这里通过`DescribeTaskStatus`接口查询对应CLB创建任务状态，该接口可以明确创建失败。
		具体实现参考`createClbPollingHandler`中 `Poll`和`Done`方法的实现。
	*/

	respPoller := poller.Poller[*ZiyanAdpt, map[string]*clb.DescribeTaskStatusResponseParams, poller.BaseDoneResult]{
		Handler: &createClbPollingHandler{opt.Region},
	}

	reqID := createResp.Response.RequestId
	result, err := respPoller.PollUntilDone(t, kt, []*string{reqID}, types.NewBatchCreateClbPollerOption())
	if err != nil {
		return nil, err
	}
	if len(result.SuccessCloudIDs) == 0 {
		return nil, errf.Newf(errf.CloudVendorError,
			"no any lb being created, TencentCloudSDK RequestId: %s", cvt.PtrToVal(reqID))
	}
	return result, nil
}

func (t *ZiyanAdpt) formatCreateClbRequest(opt *typelb.TCloudZiyanCreateClbOption) *clb.CreateLoadBalancerRequest {
	req := clb.NewCreateLoadBalancerRequest()
	// 负载均衡实例的名称
	req.LoadBalancerName = opt.LoadBalancerName
	// 负载均衡实例的网络类型。OPEN：公网属性， INTERNAL：内网属性。
	req.LoadBalancerType = common.StringPtr(string(opt.LoadBalancerType))
	// 仅适用于公网负载均衡, IP版本
	if opt.AddressIPVersion == "" {
		opt.AddressIPVersion = typelb.IPV4IPVersion
	}
	req.AddressIPVersion = (*string)(cvt.ValToPtr(opt.AddressIPVersion))
	// 负载均衡后端目标设备所属的网络
	req.VpcId = opt.VpcID
	// 负载均衡实例的类型。1：通用的负载均衡实例，目前只支持传入1。
	req.Forward = common.Int64Ptr(int64(corelb.TCloudDefaultLoadBalancerType))
	// 是否支持绑定跨地域/跨Vpc绑定IP的功能
	req.SnatPro = opt.SnatPro
	// Target是否放通来自CLB的流量。开启放通（true）：只验证CLB上的安全组；不开启放通（false）：需同时验证CLB和后端实例上的安全组
	req.LoadBalancerPassToTarget = opt.LoadBalancerPassToTarget
	// 是否创建域名化负载均衡
	req.DynamicVip = opt.DynamicVip
	req.SubnetId = opt.SubnetID
	req.Vip = opt.Vip
	req.Number = opt.Number
	req.ProjectId = opt.ProjectID
	req.SlaType = opt.SlaType
	req.ClusterIds = append(req.ClusterIds, opt.ClusterIds...)
	// 用于保证请求幂等性的字符串。该字符串由客户生成，需保证不同请求之间唯一，最大值不超过64个字符。若不指定该参数则无法保证请求的幂等性。
	req.ClientToken = opt.ClientToken
	req.ClusterTag = opt.ClusterTag
	req.EipAddressId = opt.EipAddressID
	req.SlaveZoneId = opt.SlaveZoneID
	req.Egress = opt.Egress
	req.ZoneId = opt.ZoneID
	req.MasterZoneId = opt.MasterZoneID

	req.BandwidthPackageId = opt.BandwidthPackageID
	req.SnatIps = opt.SnatIps

	for _, tag := range opt.Tags {
		req.Tags = append(req.Tags, &clb.TagInfo{
			TagKey:   cvt.ValToPtr(tag.Key),
			TagValue: cvt.ValToPtr(tag.Value),
		})
	}

	// 使用默认ISP时传递空即可
	ispVal := cvt.PtrToVal(opt.VipIsp)
	if ispVal != "" && ispVal != typelb.TCloudDefaultISP {
		req.VipIsp = opt.VipIsp
	}

	if opt.InternetChargeType != nil || opt.InternetMaxBandwidthOut != nil {
		req.InternetAccessible = &clb.InternetAccessible{
			InternetChargeType:      (*string)(opt.InternetChargeType),
			InternetMaxBandwidthOut: opt.InternetMaxBandwidthOut,
			BandwidthpkgSubType:     opt.BandwidthpkgSubType,
		}
	}

	if opt.ExclusiveCluster != nil {
		req.ExclusiveCluster = &clb.ExclusiveCluster{
			L4Clusters:       opt.ExclusiveCluster.L4Clusters,
			L7Clusters:       opt.ExclusiveCluster.L7Clusters,
			ClassicalCluster: opt.ExclusiveCluster.ClassicalCluster,
		}
	}

	if cvt.PtrToVal(opt.ZhiTong) {
		req.ZhiTong = opt.ZhiTong
	}
	if len(cvt.PtrToVal(opt.TgwGroupName)) > 0 {
		req.TgwGroupName = opt.TgwGroupName
	}
	if len(opt.Zones) > 0 {
		req.Zones = cvt.SliceToPtr(opt.Zones)
	}
	return req
}

var _ poller.PollingHandler[*ZiyanAdpt,
	map[string]*clb.DescribeTaskStatusResponseParams, poller.BaseDoneResult] = new(createClbPollingHandler)

type createClbPollingHandler struct {
	region string
}

// Done CLB 创建成功状态判断
func (h *createClbPollingHandler) Done(clbStatusMap map[string]*clb.DescribeTaskStatusResponseParams) (
	bool, *poller.BaseDoneResult) {

	result := &poller.BaseDoneResult{
		SuccessCloudIDs: make([]string, 0),
		FailedCloudIDs:  make([]string, 0),
		UnknownCloudIDs: make([]string, 0),
	}

	for _, status := range clbStatusMap {
		if status.Status == nil {
			return false, nil
		}
		switch cvt.PtrToVal(status.Status) {
		case CLBTaskStatusRunning:
			// 还有任务在运行则是没有成功
			return false, nil
		case CLBTaskStatusFail:
			result.FailedCloudIDs = cvt.PtrToSlice(status.LoadBalancerIds)
		case CLBTaskStatusSuccess:
			result.SuccessCloudIDs = cvt.PtrToSlice(status.LoadBalancerIds)
		}
	}
	return true, result
}

// Poll 返回CLB创建任务结果
func (h *createClbPollingHandler) Poll(client *ZiyanAdpt, kt *kit.Kit, requestIDs []*string) (
	map[string]*clb.DescribeTaskStatusResponseParams, error) {

	taskOpt := &typelb.TCloudDescribeTaskStatusOption{Region: h.region}
	result := make(map[string]*clb.DescribeTaskStatusResponseParams)
	// 查询对应异步任务状态
	for _, reqID := range requestIDs {
		taskOpt.TaskId = cvt.PtrToVal(reqID)
		if taskOpt.TaskId == "" {
			return nil, errors.New("empty request ID")
		}
		status, err := client.CLBDescribeTaskStatus(kt, taskOpt)
		if err != nil {
			return nil, err
		}

		result[taskOpt.TaskId] = status
	}
	return result, nil
}

// CLB异步任务状态
const (
	CLBTaskStatusSuccess = 0
	CLBTaskStatusFail    = 1
	CLBTaskStatusRunning = 2
)

// CLBDescribeTaskStatus 查询异步任务状态。
// 对于非查询类的接口（创建/删除负载均衡实例、监听器、规则以及绑定或解绑后端服务等），
// 在接口调用成功后，都需要使用本接口查询任务最终是否执行成功。
// https://cloud.tencent.com/document/api/214/30683
func (t *ZiyanAdpt) CLBDescribeTaskStatus(kt *kit.Kit, opt *typelb.TCloudDescribeTaskStatusOption) (
	*clb.DescribeTaskStatusResponseParams, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "describe task status option can not be nil")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := t.clientSet.ClbClient(opt.Region)
	if err != nil {
		return nil, fmt.Errorf("init tencent cloud clb client failed, region: %s, err: %v", opt.Region, err)
	}
	req := clb.NewDescribeTaskStatusRequest()
	if opt.TaskId != "" {
		req.TaskId = cvt.ValToPtr(opt.TaskId)
	}
	if opt.DealName != "" {
		req.DealName = cvt.ValToPtr(opt.DealName)
	}

	resp, err := client.DescribeTaskStatusWithContext(kt.Ctx, req)
	if err != nil {
		logs.Errorf("tencent cloud describe task status failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}
	return resp.Response, nil
}
