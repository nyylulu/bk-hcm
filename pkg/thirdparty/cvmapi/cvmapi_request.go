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

package cvmapi

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
)

// ReqMeta cvm request meta info
type ReqMeta struct {
	Id      string `json:"id"`
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
}

// OrderCreateReq cvm create order request
type OrderCreateReq struct {
	ReqMeta `json:",inline"`
	Params  *OrderCreateParams `json:"params"`
}

// OrderCreateParams cvm create order parameters
type OrderCreateParams struct {
	Zone              string       `json:"zone"`
	DeptName          string       `json:"deptName"`
	ProductName       string       `json:"productName"`
	Business1Id       int          `json:"business1Id"`
	Business1Name     string       `json:"business1Name"`
	Business2Id       int          `json:"business2Id"`
	Business2Name     string       `json:"business2Name"`
	Business3Id       int          `json:"business3Id"`
	Business3Name     string       `json:"business3Name"`
	ProjectId         int          `json:"projectId"`
	Image             *Image       `json:"image,omitempty"`
	InstanceType      string       `json:"instanceType"`
	SystemDiskType    string       `json:"systemDiskType"`
	SystemDiskSize    int          `json:"systemDiskSize"`
	DataDisk          []*DataDisk  `json:"dataDisk,omitempty"`
	VpcId             string       `json:"vpcId"`
	SubnetId          string       `json:"subnetId"`
	AsVpcGateway      int          `json:"asVpcGateway,omitempty"`
	ApplyNum          int          `json:"applyNum"`
	PassWord          string       `json:"passWord"`
	Security          *Security    `json:"security,omitempty"`
	IsSecurityService int          `json:"isSecurityService,omitempty"`
	IsMonitorService  int          `json:"isMonitorService,omitempty"`
	RecoverGrpId      string       `json:"recoverGrpId,omitempty"`
	InstanceName      string       `json:"instanceName,omitempty"`
	UseTime           string       `json:"useTime,omitempty"`
	Memo              string       `json:"memo,omitempty"`
	Operator          string       `json:"operator"`
	BakOperator       string       `json:"bakOperator"`
	ObsProject        string       `json:"obsProject"`
	ResourceType      ResourceType `json:"resourceType,omitempty"`
	ChargeType        ChargeType   `json:"chargeType,omitempty"`
	ChargeMonths      uint         `json:"chargeMonths,omitempty"`
	InheritInstanceId string       `json:"inheritInstanceId,omitempty"`
}

// ResourceType 申请类型
type ResourceType int

const (
	// ResourceTypeNormal 常规申领
	ResourceTypeNormal ResourceType = 0
	// ResourceTypeQuick 小额快速申领
	ResourceTypeQuick ResourceType = 5
	// ResourceTypeMachineFamily 机型族申领
	ResourceTypeMachineFamily ResourceType = 7
)

// ChargeType charge type
type ChargeType string

// ChargeType charge type
const (
	// ChargeTypePrePaid 计费模式:包年包月
	ChargeTypePrePaid ChargeType = "PREPAID"
	// ChargeTypePostPaidByHour 计费模式:按量计费
	ChargeTypePostPaidByHour ChargeType = "POSTPAID_BY_HOUR"
)

// Validate 计费模式校验
func (ct ChargeType) Validate() error {
	switch ct {
	case ChargeTypePrePaid, ChargeTypePostPaidByHour:
		return nil
	default:
		return fmt.Errorf("charge_type invalid value: %s", ct)
	}
}

// Image cvm image specification
type Image struct {
	ImageId   string `json:"imageId"`
	ImageName string `json:"imageName"`
	ImageOs   string `json:"imageOs,omitempty"`
	ImageType string `json:"imageType,omitempty"`
}

// DataDisk cvm specification
type DataDisk struct {
	DataDiskType enumor.DiskType `json:"dataDiskType"`
	DataDiskSize int             `json:"dataDiskSize"`
}

// Security cvm security specification
type Security struct {
	SecurityGroupId   string `json:"securityGroupId"`
	SecurityGroupName string `json:"securityGroupName"`
	SecurityGroupDesc string `json:"securityGroupDesc"`
}

// OrderQueryReq cvm order query request
type OrderQueryReq struct {
	ReqMeta `json:",inline"`
	Params  *OrderQueryParam `json:"params"`
}

// OrderQueryParam cvm order query parameters
type OrderQueryParam struct {
	OrderId []string `json:"orderId,omitempty"`
	// optional, query orders with certain status
	Status []int `json:"status,omitempty"`
}

// InstanceQueryReq cvm instance query request
type InstanceQueryReq struct {
	ReqMeta `json:",inline"`
	Params  *InstanceQueryParam `json:"params"`
}

// InstanceQueryParam cvm instance query parameters
type InstanceQueryParam struct {
	OrderId    []string `json:"orderId,omitempty"`
	InstanceId []string `json:"instanceId,omitempty"`
	LanIp      []string `json:"lanIp,omitempty"`
	AssetId    []string `json:"instanceAssetId,omitempty"`
}

// PlanOrderChangeReq cvm and cbs plan order change request
type PlanOrderChangeReq struct {
	ReqMeta `json:",inline"`
	Params  *PlanOrderChangeParam `json:"params"`
}

type PlanOrderChangeParam struct {
	Page            *Page    `json:"page"`
	Period          *Period  `json:"period,omitempty"`
	OrderId         []string `json:"orderId,omitempty"`
	BgName          []string `json:"bgName,omitempty"`
	DeptName        []string `json:"deptName,omitempty"`
	PlanProductName []string `json:"planProductName,omitempty"`
	InstanceFamily  []string `json:"instanceFamily,omitempty"`
	InstanceType    []string `json:"instanceType,omitempty"`
	ProjectName     []string `json:"projectName,omitempty"`
	// ResourceMode 资源模式（按机型/按机型族）
	ResourceMode string   `json:"resourceMode,omitempty"`
	UseTime      *UseTime `json:"useTime,omitempty"`
	CityName     []string `json:"cityName,omitempty"`
	ZoneName     []string `json:"zoneName,omitempty"`
	InPlan       bool     `json:"inPlan,omitempty"`
}

// DemandChangeLogQueryReq cvm and cbs demand change log query request
type DemandChangeLogQueryReq struct {
	ReqMeta `json:",inline"`
	Params  *DemandChangeLogQueryParam `json:"params"`
}

// DemandChangeLogQueryParam cvm and cbs demand change log query parameters
type DemandChangeLogQueryParam struct {
	Page         *Page   `json:"page"`
	DemandIdList []int64 `json:"demandIdList,omitempty"`
}

// CvmCbsPlanPenaltyRatioReportReq cvm and cbs plan penalty ratio report request
type CvmCbsPlanPenaltyRatioReportReq struct {
	ReqMeta `json:",inline"`
	Params  *CvmCbsPlanPenaltyRatioReportParam `json:"params"`
}

// CvmCbsPlanPenaltyRatioReportParam cvm and cbs plan penalty ratio report parameters
type CvmCbsPlanPenaltyRatioReportParam struct {
	YearMonth string                   `json:"yearMonth"`
	Data      []CvmCbsPlanProductRatio `json:"data"`
}

// CvmCbsPlanProductRatio cvm and cbs plan, plan product penalty ratio parameters
type CvmCbsPlanProductRatio struct {
	GroupDeptId           []int64         `json:"groupDeptId"`
	GroupPlanProductId    []int64         `json:"groupPlanProductId"`
	ProductIdPartitionMap map[int64]int64 `json:"productIdPartitionMap"`
	Memo                  string          `json:"memo,omitempty"`
}

// CvmCbsPlanQueryReq cvm and cbs plan info query request
type CvmCbsPlanQueryReq struct {
	ReqMeta `json:",inline"`
	Params  *CvmCbsPlanQueryParam `json:"params"`
}

// CvmCbsPlanQueryParam cvm and cbs plan info query parameters
type CvmCbsPlanQueryParam struct {
	Page            *Page    `json:"page"`
	Period          *Period  `json:"period,omitempty"`
	UseTime         *UseTime `json:"useTime,omitempty"`
	OrderIdList     []string `json:"orderIdList,omitempty"`
	DemandIdList    []int64  `json:"demandIdList,omitempty"`
	BgName          []string `json:"bgName,omitempty"`
	DeptName        []string `json:"deptName,omitempty"`
	InstanceType    []string `json:"instanceType,omitempty"`
	PlanProductName []string `json:"planProductName,omitempty"`
	ProjectName     []string `json:"projectName,omitempty"`
	CityName        []string `json:"cityName,omitempty"`
	ZoneName        []string `json:"zoneName,omitempty"`
	NotNeedWeekType bool     `json:"notNeedWeekType,omitempty"`
	UserName        string   `json:"userName,omitempty"`
}

// Page restrict the returned start index and returned number of plan items for cvm&cbs planinfo query
type Page struct {
	Start int `json:"start"`
	Size  int `json:"size"`
}

// Period restrict the submit month of plan items, format yyyy-MM for cvm&cbs planinfo query
type Period struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// UseTime -- restrict the use time of plan items,format yyyy-MM-dd for cvm&cbs planinfo query
type UseTime struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// CvmCbsPlanAdjustReq cvm and cbs plan info adjust request
type CvmCbsPlanAdjustReq struct {
	ReqMeta `json:",inline"`
	Params  *CvmCbsPlanAdjustParam `json:"params"`
}

// CvmCbsPlanAdjustParam cvm and cbs plan info adjust parameters
type CvmCbsPlanAdjustParam struct {
	BaseInfo    *AdjustBaseInfo      `json:"baseInfo"`
	SrcData     []*AdjustSrcData     `json:"srcData"`
	UpdatedData []*AdjustUpdatedData `json:"updatedData"`
	UserName    string               `json:"userName"`
}

// AdjustBaseInfo adjust base info for cvm and cbs plan info adjust params
type AdjustBaseInfo struct {
	DeptId          int    `json:"deptId"`
	DeptName        string `json:"deptName"`
	PlanProductName string `json:"planProductName"`
	Desc            string `json:"desc"`
}

// AdjustSrcData adjust source data for cvm and cbs plan info adjust params
type AdjustSrcData struct {
	AdjustType           string `json:"adjustType"`
	*CvmCbsPlanQueryItem `json:",inline"`
}

// AdjustUpdatedData adjust target data for cvm and cbs plan info adjust params
type AdjustUpdatedData struct {
	AdjustType           string  `json:"adjustType"`
	TimeAdjustCvmAmount  float32 `json:"timeAdjustCvmAmount,omitempty"`
	*CvmCbsPlanQueryItem `json:",inline"`
}

// AddCvmCbsPlanReq add cvm and cbs plan order request
type AddCvmCbsPlanReq struct {
	ReqMeta `json:",inline"`
	Params  *AddCvmCbsPlanParam `json:"params"`
}

// AddCvmCbsPlanParam add cvm and cbs plan order parameters
type AddCvmCbsPlanParam struct {
	Operator string         `json:"operator"`
	DeptName string         `json:"deptName"`
	Desc     string         `json:"desc"`
	Items    []*AddPlanItem `json:"items"`
}

/*
{
    "id":"1",
    "jsonrpc":"2.0",
    "method":"addYuntiOrder",
    "params":{
        "operator":"dommyzhang",
        "deptName":"IEG技术运营部",
        "items":[
            {
                "useTime":"2022-10-12",
                "projectName":"机房裁撤",
                "planProductName":"互娱运营支撑产品",
                "cityName":"上海",
                "zoneName":"上海五区",
                "coreTypeName":"小核心",
                "instanceModel":"S5.2XLARGE16",
                "cvmAmount":0,
                "coreAmount":200,
                "desc":"",
                "instanceIO":15,
                "diskTypeName":"高性能云硬盘",
                "diskAmount":60000
            }
        ]
    }
}
*/

// AddPlanItem add cvm and cbs plan order item
type AddPlanItem struct {
	UseTime         string  `json:"useTime"`
	ProjectName     string  `json:"projectName"`
	PlanProductName string  `json:"planProductName"`
	CityName        string  `json:"cityName"`
	ZoneName        string  `json:"zoneName"`
	CoreTypeName    string  `json:"coreTypeName"`
	InstanceModel   string  `json:"instanceModel"`
	CvmAmount       float64 `json:"cvmAmount"`
	CoreAmount      int     `json:"coreAmount"`
	Desc            string  `json:"desc"`
	InstanceIO      int     `json:"instanceIO"`
	DiskTypeName    string  `json:"diskTypeName"`
	DiskAmount      int     `json:"diskAmount"`
}

// QueryPlanOrderReq query cvm and cbs plan order request
type QueryPlanOrderReq struct {
	ReqMeta `json:",inline"`
	Params  *QueryPlanOrderParam `json:"params"`
}

// QueryPlanOrderParam query cvm and cbs plan order parameters
type QueryPlanOrderParam struct {
	OrderIds []string `json:"orderIds,omitempty"`
}

/* CapacityReq request example
{
    "method":"queryApplyCapacity",
    "params":{
        "deptId":1041,
        "type":2,
        "business3Id":1388520,
        "cloudCampus":"ap-guangzhou-4",
        "instanceType":"S2.SMALL2",
        "vpcId":"vpc-rd18ho77",
        "subnetId":"subnet-6ka02gb6",
        "projectName":"常规项目",
        "systemDiskInfo":{
            "systemDiskType":"CLOUD_PREMIUM",
            "systemDiskSize":100
        },
        "dataDiskInfo":[

        ],
        "ResourceType":0
    },
    "jsonrpc":"2.0",
    "id":"16477579036464836"
}
*/

// CapacityReq cvm capacity query request
type CapacityReq struct {
	ReqMeta `json:",inline"`
	Params  *CapacityParam `json:"params"`
}

// CapacityParam cvm capacity query parameters
type CapacityParam struct {
	DeptId         int             `json:"deptId"`
	Business3Id    int             `json:"business3Id"`
	CloudCampus    string          `json:"cloudCampus"`
	InstanceType   string          `json:"instanceType"`
	VpcId          string          `json:"vpcId"`
	SubnetId       string          `json:"subnetId"`
	ProjectName    string          `json:"projectName"`
	ChargeType     ChargeType      `json:"chargeType"`
	SystemDiskInfo *SysDiskInfo    `json:"systemDiskInfo,omitempty"`
	DataDiskInfo   []*DataDiskInfo `json:"dataDiskInfo,omitempty"`
}

// SysDiskInfo system disk info
type SysDiskInfo struct {
	SystemDiskType string `json:"systemDiskType"`
	SystemDiskSize int    `json:"systemDiskSize,omitempty"`
}

// DataDiskInfo data disk info
type DataDiskInfo struct {
	SystemDiskType string `json:"dataDiskType"`
	SystemDiskSize int    `json:"dataDiskSize,omitempty"`
}

// VpcReq cvm vpc query request
type VpcReq struct {
	ReqMeta `json:",inline"`
	Params  *VpcParam `json:"params"`
}

// VpcParam cvm vpc query parameters
type VpcParam struct {
	DeptId int    `json:"deptId"`
	Region string `json:"region"`
}

// SubnetReq cvm subnet query request
type SubnetReq struct {
	ReqMeta `json:",inline"`
	Params  *SubnetParam `json:"params"`
}

// SubnetParam cvm subnet query parameters
type SubnetParam struct {
	DeptId int    `json:"deptId"`
	Region string `json:"region"`
	Zone   string `json:"zone"`
	VpcId  string `json:"vpcId"`
}

// ReturnReq create cvm return order request
type ReturnReq struct {
	ReqMeta `json:",inline"`
	Params  *ReturnParam `json:"params"`
}

// ReturnParam create cvm return order parameters
type ReturnParam struct {
	// 选填，是否立刻销毁 0否1是, 默认0
	IsReturnNow int `json:"isReturnNow"`
	// 要退还的实例ID列表，必填
	InstanceList []string `json:"instanceList"`
	// 选填，是否同时销毁数据盘,0否1是, 默认0
	IsWithDataDisks int `json:"isWithDataDisks"`
	// 选填， 销毁类型：0-直接销毁 1-置换销毁, 默认0
	ReturnType int `json:"returnType"`
	// 选填，如果是置换销毁，填写原因
	Reason string `json:"reason"`
	// 选填， 退回预算项目。默认常规项目
	ObsProject string `json:"obsProject"`
	// 选填，是否强制销毁，默认为false，默认情况下会校验进程端口绑定情况，对于校验不通过的设备禁止提销毁单
	Force bool `json:"force"`
	// 选填，是否接受成本分摊。true是，false否。默认：false
	AcceptCostShare bool `json:"acceptCostShare"`
}

// GetCvmProcessReq get cvm process request
type GetCvmProcessReq struct {
	ReqMeta `json:",inline"`
	Params  *GetCvmProcessParam `json:"params"`
}

// GetCvmProcessParam get cvm process parameters
type GetCvmProcessParam struct {
	AssetIds []string `json:"instanceAssetId"`
}

// GetErpProcessReq get erp process request
type GetErpProcessReq struct {
	ReqMeta `json:",inline"`
	Params  *GetErpProcessParam `json:"params"`
}

// GetErpProcessParam get erp process parameters
type GetErpProcessParam struct {
	AssetIds []string `json:"logicPcCode"`
}

// ReturnDetailReq query cvm return order detail request
type ReturnDetailReq struct {
	ReqMeta `json:",inline"`
	Params  *ReturnDetailParam `json:"params"`
}

// ReturnDetailParam query cvm return order detail parameters
type ReturnDetailParam struct {
	OrderId string `json:"orderId"`
	Page    *Page  `json:"page,omitempty"`
}

// QueryCvmInstanceTypeReq query cvm instance type request
type QueryCvmInstanceTypeReq struct {
	ReqMeta `json:",inline"`
	Params  *QueryCvmInstanceTypeParams `json:"params"`
}

// QueryCvmInstanceTypeParams query cvm instance type parameters
type QueryCvmInstanceTypeParams struct {
	InstanceClass []string `json:"instanceClass,omitempty"`
	InstanceType  []string `json:"instanceType,omitempty"`
	InstanceGroup []string `json:"instanceGroup,omitempty"`
}

// GetApproveLogReq get approve log request
type GetApproveLogReq struct {
	ReqMeta `json:",inline"`
	Params  *GetApproveLogParams `json:"params"`
}

// GetApproveLogParams get approve log parameters
type GetApproveLogParams struct {
	OrderId []string `json:"orderId"`
}

// GetCvmApproveLogReq get cvm approve log request
type GetCvmApproveLogReq struct {
	ReqMeta `json:",inline"`
	Params  *GetCvmApproveLogParams `json:"params"`
}

// GetCvmApproveLogParams get cvm approve log parameters
type GetCvmApproveLogParams struct {
	OrderId string `json:"orderId"`
}

// RevokeCvmOrderReq ...
type RevokeCvmOrderReq struct {
	ReqMeta `json:",inline"`
	Params  *RevokeCvmOrderParams `json:"params"`
}

// RevokeCvmOrderParams ...
type RevokeCvmOrderParams struct {
	OrderId string `json:"order_id"`
}
