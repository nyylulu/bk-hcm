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

// Package cvmapi ...
package cvmapi

import (
	"strconv"
	"time"
)

var mapObsProject = map[int64]string{
	1: "常规项目",
	2: "春节保障",
	3: "机房裁撤",
	4: "常规项目",
	5: "短租项目",
	6: "滚服项目",
}

// GetObsProject get OBS project by CR require type.
func GetObsProject(requireType int64) string {
	switch requireType {
	case 1, 4, 5, 6:
		return mapObsProject[requireType]
	case 2:
		return getSpringObsProject()
	case 3:
		return getDissolveObsProject()
	default:
		// return "常规项目" as default
		return mapObsProject[1]
	}
}

func getSpringObsProject() string {
	// 春保窗口期：12月1日～次年3月15日
	// 12月1日～12月31日提单的春保项目前缀为次年
	year := time.Now().Local().Year()
	if time.Now().Month() == time.December {
		year += 1
	}

	prefixYear := strconv.Itoa(year)
	project := prefixYear + "春节保障"

	return project
}

func getDissolveObsProject() string {
	// TODO:
	// 暂定按自然年作为机房裁撤的窗口滚动周期
	// 如"2024机房裁撤"
	year := time.Now().Local().Year()
	prefixYear := strconv.Itoa(year)
	project := prefixYear + "机房裁撤"

	return project
}

// CvmCbsPlanModityType 需求预测接口调整类型
var CvmCbsPlanModityType = map[int64]string{
	1: "add",
	2: "delete",
	3: "update",
}

const (
	// CvmId CVM请求ID
	CvmId = "1"
	// CvmJsonRpc CVM请求JSONRPC
	CvmJsonRpc = "2.0"

	// CvmDeptId CVM容量查询部门ID
	CvmDeptId = 1041
	// CvmLaunchDeptName CVM生产时部门名称
	CvmLaunchDeptName = "IEG技术运营部"
	// CvmLaunchProductName CVM运营产品名（项目名）
	CvmLaunchProductName = "互娱资源公共平台"
	// CvmLaunchBiz1Id CVM一级业务ID
	CvmLaunchBiz1Id = 656545
	// CvmLaunchBiz1Name CVM一级业务名
	CvmLaunchBiz1Name = "CC_资源运营服务"
	// CvmLaunchBiz2Id CVM二级业务ID
	CvmLaunchBiz2Id = 656560
	// CvmLaunchBiz2Name CVM二级业务名
	CvmLaunchBiz2Name = "CC_资源运营服务"
	// CvmLaunchBiz3Id CVM三级业务ID
	CvmLaunchBiz3Id = 1073015
	// CvmLaunchBiz3Name CVM三级业务名
	CvmLaunchBiz3Name = "CC_SCR_加工池"
	// CvmLaunchSystemDiskTypePremium CVM生产时系统盘类型，当前固定为高性能云盘
	CvmLaunchSystemDiskTypePremium = "CLOUD_PREMIUM"
	// CvmLaunchSystemDiskTypeBasic CVM生产时系统盘类型，对于固定为本地盘
	CvmLaunchSystemDiskTypeBasic = "LOCAL_BASIC"
	// CvmLaunchSystemDiskSizePremium CVM生产时系统盘大小，当前固定为100G
	CvmLaunchSystemDiskSizePremium = 100
	// CvmLaunchSystemDiskSizeBasic CVM生产时系统盘大小，对于IT设备固定为50G
	CvmLaunchSystemDiskSizeBasic = 50
	// CVM_LAUNCH_USETIME CVM生产时数据盘类型，当前固定为高性能云盘
	CVM_LAUNCH_USETIME = "0000-00-00 00:00:00" // CVM生产时必填项，yuti开发对该字段含义也未知，暂时写死为该固定值0000-00-00 00:00:00
	// CvmLaunchProjectId CVM项目ID，yuti开发对该字段含义也未知，暂时写死为固定值0
	CvmLaunchProjectId = 0
	// CvmOrderLinkPrefix CVM生产单据详情链接前缀
	CvmOrderLinkPrefix = "https://yunti.woa.com/orders/cvm/"
	// CvmReturnLinkPrefix CVM退回单据详情链接前缀
	CvmReturnLinkPrefix = "https://yunti.woa.com/orders/cvmreturn/"
	// CvmPlanLinkPrefix CVM&CBS需求单据详情链接前缀
	CvmPlanLinkPrefix = "https://yunti.woa.com/orders/iaasplan/"

	// CvmSeparateCampus 分Campus
	CvmSeparateCampus = "cvm_separate_campus"

	// CvmApiKey CVM API key
	CvmApiKey = "api_key"
	// CvmApiKeyVal CVM API key value
	CvmApiKeyVal = "octopuskg"

	// CvmCbsPlanQueryId 需求预测查询id
	CvmCbsPlanQueryId = "16318853269804145"
	// CvmCbsPlanAdjustId 需求预测调整id
	CvmCbsPlanAdjustId = "16319322822855177"

	// CvmCbsPlanQueryBgName 需求预测首页查询接口事业群名称
	CvmCbsPlanQueryBgName = "IEG互动娱乐事业群"
	// CvmCbsPlanDeptId 需求预测接口事业群ID
	CvmCbsPlanDeptId = 1041
	// DefaultPlanProductName 需求预测默认规划产品
	DefaultPlanProductName = "互娱运营支撑产品"

	// CvmLaunchMethod cvm methods
	// 创建CVM订单方法
	CvmLaunchMethod = "createCvmOrder"
	// CvmOrderStatusMethod CVM单据进度查询方法
	CvmOrderStatusMethod = "queryOrders"
	// CvmInstanceStatusMethod CVM实例状态查询方法
	CvmInstanceStatusMethod = "queryCVMInstances"
	// CvmCapacityMethod CVM容量查询方法
	CvmCapacityMethod = "queryApplyCapacity"
	// CvmVpcMethod CVM vpc信息查询方法
	CvmVpcMethod = "getVpcInfo"
	// CvmSubnetMethod CVM subnet信息查询方法
	CvmSubnetMethod = "getSubNetInfo"
	// CvmCbsDemandChangeLogQueryMethod 预测需求的变更记录查询接口
	CvmCbsDemandChangeLogQueryMethod = "queryDemandChangeLogForIEG"
	// CvmCbsPlanQueryMethod 需求预测首页查询接口
	CvmCbsPlanQueryMethod = "queryCvmCbsInfoForIEG"
	// CvmCbsPlanAdjustMethod 需求预测首页调整接口
	CvmCbsPlanAdjustMethod = "adjustOrder"
	// CvmCbsPlanAutoAdjustMethod 需求预测细粒度调整接口
	CvmCbsPlanAutoAdjustMethod = "submitAutoAdjustOrder"
	// CvmCbsPlanAddMethod 需求预测追加接口
	CvmCbsPlanAddMethod = "addYuntiOrder"
	// CvmCbsPlanOrderQueryMethod 需求单据查询接口
	CvmCbsPlanOrderQueryMethod = "queryYuntiOrder"
	// CvmGetProcessMethod CVM流程查询方法
	CvmGetProcessMethod = "getCVMProcess"
	// GetErpProcessMethod ERP流程查询方法
	GetErpProcessMethod = "getERPProcess"
	// CvmReturnMethod CVM退回提单方法
	CvmReturnMethod = "createCvmReturnOrder"
	// CvmReturnStatusMethod CVM退回单据状态查询方法
	CvmReturnStatusMethod = "queryCvmReturnOrder"
	// CvmReturnDetailMethod 根据单号查询退回CVM方法
	CvmReturnDetailMethod = "queryReturnCvmByOrder"
	// QueryCvmInstanceType 查询CVM机型信息
	QueryCvmInstanceType = "queryCvmInstanceType"

	// CvmCbsPlanDefaultDesc 需求预测单据的默认备注，用于管理员判断需求来源
	CvmCbsPlanDefaultDesc = "[From IEG HCM]"

	// DftImageID default image id of TencentOS Server 2.6 (TK4)
	DftImageID = "img-fjxtfi0n"

	// AdjustTypeAdjust 预测调整类型-常规修改
	AdjustTypeAdjust = "常规修改"
	// AdjustTypeDelay 预测调整类型-加急延期
	AdjustTypeDelay = "加急延期"
	// AdjustTypeCancel 预测调整类型-需求取消
	AdjustTypeCancel = "需求取消"
)

// CVMCli yunti client options
type CVMCli struct {
	// CvmApiAddr yunti api address
	CvmApiAddr        string `yaml:"host"`
	CvmLaunchPassword string `yaml:"launch_password"`
}
