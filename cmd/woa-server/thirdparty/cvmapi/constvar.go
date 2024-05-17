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
	"strconv"
	"time"
)

var mapObsProject = map[int64]string{
	1: "常规项目",
	2: "春节保障",
	3: "机房裁撤",
	4: "常规项目",
	5: "短租项目",
}

// GetObsProject get OBS project by CR require type.
func GetObsProject(requireType int64) string {
	switch requireType {
	case 1, 4, 5:
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

// 需求预测接口调整类型
var CvmCbsPlanModityType = map[int64]string{
	1: "add",
	2: "delete",
	3: "update",
}

const (
	// CVM请求ID
	CvmId = "1"
	// CVM请求JSONRPC
	CvmJsonRpc = "2.0"

	// CVM容量查询部门ID
	CvmDeptId = 1041
	// CVM生产时部门名称
	CvmLaunchDeptName = "IEG技术运营部"
	// CVM运营产品名（项目名）
	CvmLaunchProductName = "互娱资源公共平台"
	// CVM一级业务ID
	CvmLaunchBiz1Id = 656545
	// CVM一级业务名
	CvmLaunchBiz1Name = "CC_资源运营服务"
	// CVM二级业务ID
	CvmLaunchBiz2Id = 656560
	// CVM二级业务名
	CvmLaunchBiz2Name = "CC_资源运营服务"
	// CVM三级业务ID
	CvmLaunchBiz3Id = 1073015
	// CVM三级业务名
	CvmLaunchBiz3Name = "CC_SCR_加工池"
	// CVM生产时系统盘类型，当前固定为高性能云盘
	CvmLaunchSystemDiskTypePremium = "CLOUD_PREMIUM"
	// CVM生产时系统盘类型，对于固定为本地盘
	CvmLaunchSystemDiskTypeBasic = "LOCAL_BASIC"
	// CVM生产时系统盘大小，当前固定为100G
	CvmLaunchSystemDiskSizePremium = 100
	// CVM生产时系统盘大小，对于IT设备固定为50G
	CvmLaunchSystemDiskSizeBasic = 50
	CVM_LAUNCH_USETIME           = "0000-00-00 00:00:00" //CVM生产时必填项，yuti开发对该字段含义也未知，暂时写死为该固定值0000-00-00 00:00:00
	// CVM项目ID，yuti开发对该字段含义也未知，暂时写死为固定值0
	CvmLaunchProjectId = 0
	// CVM生产密码
	// TODO: IMPORTANT!!! get from config
	CvmLaunchPassword = "bG5T2OTx3rP6" //CVM生产密码
	// CVM生产单据详情链接前缀
	CvmOrderLinkPrefix = "https://yunti.woa.com/orders/cvm/"
	// CVM退回单据详情链接前缀
	CvmReturnLinkPrefix = "https://yunti.woa.com/orders/cvmreturn/"
	// CVM&CBS需求单据详情链接前缀
	CvmPlanLinkPrefix = "https://yunti.woa.com/orders/iaasplan/"

	// 分Campus
	CvmSeparateCampus = "cvm_separate_campus"

	// CVM API key
	CvmApiKey = "api_key"
	// CVM API key value
	CvmApiKeyVal = "octopuskg"

	// 需求预测查询id
	CvmCbsPlanQueryId = "16318853269804145"
	// 需求预测调整id
	CvmCbsPlanAdjustId = "16319322822855177"

	// 需求预测首页查询接口事业群名称
	CvmCbsPlanQueryBgName = "IEG互动娱乐事业群"
	// 需求预测接口事业群ID
	CvmCbsPlanDeptId = 1041
	// 需求预测默认规划产品
	DefaultPlanProductName = "互娱运营支撑产品"

	// cvm methods
	// 创建CVM订单方法
	CvmLaunchMethod = "createCvmOrder"
	// CVM单据进度查询方法
	CvmOrderStatusMethod = "queryOrders"
	// CVM实例状态查询方法
	CvmInstanceStatusMethod = "queryCVMInstances"
	// CVM容量查询方法
	CvmCapacityMethod = "queryApplyCapacity"
	// CVM vpc信息查询方法
	CvmVpcMethod = "getVpcInfo"
	// CVM subnet信息查询方法
	CvmSubnetMethod = "getSubNetInfo"
	// 需求预测首页查询接口
	CvmCbsPlanQueryMethod = "queryCvmCbsInfo"
	// 需求预测首页调整接口
	CvmCbsPlanAdjustMethod = "adjustOrder"
	// 需求预测追加接口
	CvmCbsPlanAddMethod = "addYuntiOrder"
	// 需求单据查询接口
	CvmCbsPlanOrderQueryMethod = "queryYuntiOrder"
	// CVM流程查询方法
	CvmGetProcessMethod = "getCVMProcess"
	// ERP流程查询方法
	GetErpProcessMethod = "getERPProcess"
	// CVM退回提单方法
	CvmReturnMethod = "createCvmReturnOrder"
	// CVM退回单据状态查询方法
	CvmReturnStatusMethod = "queryCvmReturnOrder"
	// 根据单号查询退回CVM方法
	CvmReturnDetailMethod = "queryReturnCvmByOrder"

	// DftImageID default image id of TencentOS Server 2.6 (TK4)
	DftImageID = "img-fjxtfi0n"
)
