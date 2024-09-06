/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package erpapi ...
package erpapi

// ErpResp erp response
type ErpResp struct {
	DataSet ErpDataSet `json:"dataSet"`
}

// ErpDataSet erp response data set
type ErpDataSet struct {
	Header RespMeta    `json:"header"`
	Data   interface{} `json:"data"`
}

// RespMeta erp response meta info
type RespMeta struct {
	Version string `json:"version"`
	ErrMsg  string `json:"errorInfo"`
	Code    int    `json:"returnCode"`
}

// ReturnRespData device return response data
type ReturnRespData struct {
	OrderId string `json:"returnOrderId"`
}

// OrderQueryRespData query return order response data
type OrderQueryRespData struct {
	ResultSet []*OrderQueryRst `json:"resultSet"`
}

// OrderQueryRst query return order result item
type OrderQueryRst struct {
	OrderId string `json:"orderId"`
	/*
		init_status = 0    //初始
		allow_checking_status = 1  //准入检查中
		allow_check_pass_status = 2   //准入检查通过
		allow_check_reject_status = 3//准入检查驳回
		checking_status =4  //设备检查中
		check_pass_status = 5 //设备检查通过
		check_reject_status = 6 //设备检查驳回
		approving_status = 7  //审核中
		approve_pass_status = 8  //审核通过
		approve_reject_status = 9 //审核驳回
		recycling_status = 10  //回收中
		erp_recycling_status = 14  //回收中
		recycle_success_status = 11  //回收完成
		recycle_failure_status = 12  //回收失败
	*/
	Status int `json:"status"`
	/*
	   init_status = 0  //初始
	   allow_checking_status =1    //准入检查中
	   allow_check_pass_status =2 //准入检查通过
	   allow_check_reject_status =3  //准入检查驳回
	   checking_status = 4 //设备检查中
	   check_finish_status = 5 //设备检查完成
	*/
	CheckStatus int `json:"checkStatus"`
	// RecycleStatus return status
	// 0: 等待回收;
	// 4: 下发ERP回收中;
	// 5: ERP回收处理中;
	// 6: 下发ERP回收失败(重复退回);
	// 7: 设备回收成功;
	// 8: 设备回收失败，待人工处理;
	// 9: ERP回收处理中
	RecycleStatus int    `json:"recycleStatus"`
	AssetId       string `json:"assetId"`
	OBSLabel      string `json:"OBSLabel"`
	CurApprover   string `json:"curApprover"`
}
