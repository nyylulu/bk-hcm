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

// RespMeta cvm response meta info
type RespMeta struct {
	Id      string    `json:"id"`
	JsonRpc string    `json:"jsonrpc"`
	TraceId string    `json:"x_trace_id"`
	Error   RespError `json:"error"`
}

// RespError cvm response error info
type RespError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OrderCreateResp cvm create order response
type OrderCreateResp struct {
	RespMeta `json:",inline"`
	Result   *OrderCreateRst `json:"result"`
}

// OrderCreateRst cvm create order result
type OrderCreateRst struct {
	OrderId string `json:"orderId"`
	Status  int    `json:"status"`
}

// OrderQueryResp cvm order query response
type OrderQueryResp struct {
	RespMeta `json:",inline"`
	Result   *OrderQueryRst `json:"result"`
}

// OrderQueryRst cvm order query result
type OrderQueryRst struct {
	Total int          `json:"total"`
	Data  []*OrderItem `json:"data"`
}

// OrderItem cvm order info
type OrderItem struct {
	OrderId string `json:"orderId"`
	// 单据状态：
	// 8完成
	// 0待部门管理员审批,1待业务总监审批,2待规划经理审批,3待资源审批,4待生成CDH宿主机,
	// 5CDH宿主机生成中,6待生成CVM,7CVM生成中,127驳回,129下发生产失败
	Status int `json:"status"`
}

const (
	OrderStatusFinish int = 8
	OrderStatusReject int = 127
	OrderStatusFailed int = 129
)

// InstanceQueryResp cvm instance query response
type InstanceQueryResp struct {
	RespMeta `json:",inline"`
	Result   *InstanceQueryRst `json:"result"`
}

// InstanceQueryResp cvm instance query result
type InstanceQueryRst struct {
	Total int             `json:"total"`
	Data  []*InstanceItem `json:"data"`
}

// InstanceItem cvm instance info
type InstanceItem struct {
	InstanceId      string `json:"instanceId"`
	InstanceStatus  string `json:"instanceStatus"`
	AssetId         string `json:"instanceAssetId"`
	LanIp           string `json:"lanIp"`
	WanIp           string `json:"wanIp"`
	OwnerLanIp      string `json:"ownerLanIp"`
	CloudCampus     string `json:"cloudCampus"`
	SecurityGroupId string `json:"securityGroupId"`
	ImageName       string `json:"imageName"`
	PrivateVpcId    string `json:"privateVpcId"`
	CloudRegion     string `json:"cloudRegion"`
	PrivateSubnetId string `json:"privateSubnetId"`
	CreateTime      string `json:"createTime"`
	Pool            int    `json:"pool"`
	ObsProject      string `json:"obsProject"`
}

// CvmCbsPlanQueryResp cvm and cbs plan query response
type CvmCbsPlanQueryResp struct {
	RespMeta  `json:",inline"`
	Result    *CvmCbsPlanQueryRst `json:"result"`
	Errorinfo interface{}         `json:"errorinfo"`
}

// CvmCbsPlanQueryRst cvm and cbs plan query result
type CvmCbsPlanQueryRst struct {
	Total         int                    `json:"total"`
	Data          []*CvmCbsPlanQueryItem `json:"data"`
	AllCvmAmount  float32                `json:"allCvmAmount"`
	AllCoreAmount float32                `json:"allCoreAmount"`
}

// CvmCbsPlanQueryItem cvm and cbs plan query item
type CvmCbsPlanQueryItem struct {
	BaseCoreAmount        int     `json:"baseCoreAmount"`
	BaseCvmAmount         float32 `json:"baseCvmAmount"`
	SliceId               string  `json:"sliceId"`
	YearMonth             string  `json:"yearMonth"`
	Year                  int     `json:"year"`
	Month                 int     `json:"month"`
	UseTime               string  `json:"useTime"`
	BgId                  int     `json:"bgId"`
	BgName                string  `json:"bgName"`
	DeptId                int     `json:"deptId"`
	DeptName              string  `json:"deptName"`
	PlanProductId         int     `json:"planProductId"`
	PlanProductName       string  `json:"planProductName"`
	ProjectName           string  `json:"projectName"`
	OrderId               string  `json:"orderId"`
	CityId                int     `json:"cityId"`
	CityName              string  `json:"cityName"`
	ZoneId                int     `json:"zoneId"`
	ZoneName              string  `json:"zoneName"`
	CoreType              int     `json:"coreType"`
	CoreTypeName          string  `json:"coreTypeName"`
	InstanceType          string  `json:"instanceType"`
	InstanceModel         string  `json:"instanceModel"`
	InstanceIO            int     `json:"instanceIO"`
	DiskType              int     `json:"diskType"`
	DiskTypeName          string  `json:"diskTypeName"`
	CvmAmount             float32 `json:"cvmAmount"`
	RamAmount             float32 `json:"ramAmount"`
	CoreAmount            float32 `json:"coreAmount"`
	AllDiskAmount         float32 `json:"allDiskAmount"`
	ApplyCvmAmount        float32 `json:"applyCvmAmount"`
	ApplyRamAmount        float32 `json:"applyRamAmount"`
	ApplyCoreAmount       float32 `json:"applyCoreAmount"`
	ApplyDiskAmount       float32 `json:"applyDiskAmount"`
	PlanCvmAmount         float32 `json:"planCvmAmount"`
	PlanRamAmount         float32 `json:"planRamAmount"`
	PlanCoreAmount        float32 `json:"planCoreAmount"`
	PlanDiskAmount        float32 `json:"planDiskAmount"`
	ExpiredCvmAmount      float32 `json:"expiredCvmAmount"`
	ExpiredRamAmount      float32 `json:"expiredRamAmount"`
	ExpiredCoreAmount     float32 `json:"expiredCoreAmount"`
	ExpiredDiskAmount     float32 `json:"expiredDiskAmount"`
	RealCvmAmount         float32 `json:"realCvmAmount"`
	RealRamAmount         float32 `json:"realRamAmount"`
	RealCoreAmount        float32 `json:"realCoreAmount"`
	RealDiskAmount        float32 `json:"realDiskAmount"`
	MjOrderId             string  `json:"mjOrderId"`
	RequirementStatus     int     `json:"requirementStatus"`
	RequirementStatusName string  `json:"requirementStatusName"`
	RequirementWeekType   string  `json:"requirementWeekType"`
	IsManualWeekType      int     `json:"isManualWeekType"`
	IsInProcessing        int     `json:"isInProcessing"`
	ProcessingOrderId     string  `json:"processingOrderId"`
	DemandId              string  `json:"demandId"`
}

// CvmCbsPlanAdjustResp cvm and cbs plan adjust response
type CvmCbsPlanAdjustResp struct {
	RespMeta  `json:",inline"`
	Result    *CvmCbsPlanAdjustRst `json:"result"`
	Errorinfo interface{}          `json:"errorinfo"`
}

// CvmCbsPlanAdjustRst cvm and cbs plan adjust result
type CvmCbsPlanAdjustRst struct {
	Status  int    `json:"status"`
	OrderId string `json:"orderId"`
}

// AddCvmCbsPlanResp add cvm and cbs plan order response
type AddCvmCbsPlanResp struct {
	RespMeta `json:",inline"`
	Result   *AddCvmCbsPlanRst `json:"result"`
}

// AddCvmCbsPlanRst add cvm and cbs plan order result
type AddCvmCbsPlanRst struct {
	Status  int    `json:"status"`
	OrderId string `json:"orderId"`
}

// QueryPlanOrderResp query cvm and cbs plan order response
type QueryPlanOrderResp struct {
	RespMeta `json:",inline"`
	Result   map[string]*QueryPlanOrderRst `json:"result"`
}

// AddCvmCbsPlanRst query cvm and cbs plan order result
type QueryPlanOrderRst struct {
	Code int            `json:"code"`
	Data *PlanOrderData `json:"data"`
}

// PlanOrderData query cvm and cbs plan order data
type PlanOrderData struct {
	BaseInfo *PlanOrderBaseInfo `json:"baseInfo"`
}

type PlanOrderStatus int

const (
	// PlanOrderStatusDeptAdmin 部门管理员审批
	PlanOrderStatusDeptAdmin PlanOrderStatus = 1
	// PlanOrderStatusPlanManager 规划经理审批
	PlanOrderStatusPlanManager PlanOrderStatus = 2
	// PlanOrderStatusResManager 资源经理审批
	PlanOrderStatusResManager PlanOrderStatus = 3
	// PlanOrderStatusFinished 申请结束
	PlanOrderStatusFinished PlanOrderStatus = 4
	// PlanOrderStatusArchPlat 架平审批
	PlanOrderStatusArchPlat PlanOrderStatus = 6
	// PlanOrderStatusResGM 资源总监审批
	PlanOrderStatusResGM PlanOrderStatus = 10
	// PlanOrderStatusRejected 审批驳回
	PlanOrderStatusRejected PlanOrderStatus = 127
	// PlanOrderStatusApproved 审批通过
	PlanOrderStatusApproved PlanOrderStatus = 20
)

// PlanOrderBaseInfo query cvm and cbs plan order base info
type PlanOrderBaseInfo struct {
	Status PlanOrderStatus `json:"status"`
}

// CapacityResp cvm apply capacity query response
type CapacityResp struct {
	RespMeta `json:",inline"`
	Result   *CapacityRst `json:"result"`
}

// CapacityRst cvm apply capacity query result
type CapacityRst struct {
	MaxNum  int             `json:"maxNum"`
	MaxInfo []*CapacityInfo `json:"maxInfo"`
	Ret     int             `json:"ret"`
	Msg     string          `json:"msg"`
}

// CapacityInfo cvm apply capacity into
type CapacityInfo struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

// VpcResp cvm vpc query response
type VpcResp struct {
	RespMeta `json:",inline"`
	Result   []*VpcInfo `json:"result"`
}

// VpcRst cvm vpc query result
type VpcInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// SubnetResp cvm subnet query response
type SubnetResp struct {
	RespMeta `json:",inline"`
	Result   []*SubnetInfo `json:"result"`
}

// SubnetRst cvm subnet query result
type SubnetInfo struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	LeftIpNum int    `json:"leftIpNum"`
}

// ReturnQueryResp cvm return order query response
type ReturnQueryResp struct {
	RespMeta `json:",inline"`
	Result   *ReturnQueryRst `json:"result"`
}

// ReturnQueryRst cvm return order query result
type ReturnQueryRst struct {
	Total int                `json:"total"`
	Data  []*ReturnOrderItem `json:"data"`
}

// ReturnOrderItem cvm return order info
type ReturnOrderItem struct {
	Status      int               `json:"status"`
	Description string            `json:"statusDesc"`
	Message     string            `json:"statusMsg"`
	ReturnCnt   int               `json:"returnCount"`
	FinishCnt   int               `json:"finishCount"`
	Instances   []*ReturnInstance `json:"returnInstances"`
}

// ReturnInstance cvm return instance info
type ReturnInstance struct {
	InstanceId   string `json:"instanceId"`
	Status       int    `json:"status"`
	Summary      string `json:"summaryStatus"`
	Description  string `json:"statusDesc"`
	ReturnBudget bool   `json:"returnBudget"`
	FinishTime   string `json:"finishTime"`
}

// ReturnDetailResp cvm return order detail query response
type ReturnDetailResp struct {
	RespMeta `json:",inline"`
	Result   *ReturnDetailRst `json:"result"`
}

// ReturnDetailRst cvm return order detail query result
type ReturnDetailRst struct {
	Total int             `json:"total"`
	Data  []*ReturnDetail `json:"data"`
}

// ReturnDetail cvm return order detail info
type ReturnDetail struct {
	InstanceId string `json:"instanceId"`
	AssetId    string `json:"instanceAssetId"`
	Ip         string `json:"lanIp"`
	// instance return status:
	// 1	回收站
	// 2	云销毁中
	// 10	停用并回收IP
	// 15	从CMDB删除
	// 20	销毁完成
	// 127  审批驳回
	// 128	异常终止
	Status int `json:"status"`
	// Tag return plan tag
	Tag string `json:"tag"`
	// Partition cost sharing ratio
	Partition float64 `json:"partition"`
	// RetPlanMsg return plan and cost sharing remark
	RetPlanMsg string `json:"returnPlanMessage"`
	FinishTime string `json:"finishTime"`
}

// GetCvmProcessResp get cvm process response
type GetCvmProcessResp struct {
	RespMeta `json:",inline"`
	Result   *GetCvmProcessRst `json:"result"`
}

// GetCvmProcessRst get cvm process result
type GetCvmProcessRst struct {
	Total int               `json:"total"`
	Data  []*CvmProcessItem `json:"data"`
}

// CvmProcessItem cvm process item
type CvmProcessItem struct {
	InstanceId string `json:"instanceId"`
	AssetId    string `json:"instanceAssetId"`
	Ip         string `json:"lanIp"`
	OrderId    string `json:"orderId"`
	// StatusDesc cvm process status description
	// OTHERS(-1, "未定义的流程"),
	// EMPTY(0, ""),
	// UPGRADE(1, "升降配中"),
	// MIGRATE(2, "迁移中"),
	// EXCHANGE(8, "置换中"),
	// RETURN(9, "销毁中")
	StatusDesc string `json:"statusDesc"`
}

// GetErpProcessResp get erp process response
type GetErpProcessResp struct {
	RespMeta `json:",inline"`
	Result   *GetErpProcessRst `json:"result"`
}

// GetErpProcessRst get erp process result
type GetErpProcessRst struct {
	Total int               `json:"total"`
	Data  []*ErpProcessItem `json:"data"`
}

// ErpProcessItem erp process item
type ErpProcessItem struct {
	AssetId    string `json:"logicPcCode"`
	OrderId    string `json:"orderCode"`
	ActionType string `json:"actionType"`
}
