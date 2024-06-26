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

/*
{
    "params":{
        "content":{
            "type":"Json",
            "version":"1.0",
            "requestInfo":{
                "requestKey":"20140228001",
                "requestModule":"quota",
                "operator":"forestchen"
            },
            "requestItem":{
                "method":"deviceReturn",
                "data":{
                    "deptId":3,
                    "assetList":[
                        "TYSV16111509"
                    ],
                    "reason":"机房裁撤退回",
                    "resonMsg":"",
                    "remark":""
                }
            }
        }
    }
}
*/

// ReqInfo erp request meta info
type ReqInfo struct {
	ReqKey    string `json:"requestKey"`
	ReqModule string `json:"requestModule"`
	Operator  string `json:"operator"`
}

// ErpReq erp request
type ErpReq struct {
	Params *ErpParam `json:"params"`
}

// ErpParam erp request parameters
type ErpParam struct {
	Content *Content `json:"content"`
}

// Content erp request parameter content
type Content struct {
	Type    string   `json:"type"`
	Version string   `json:"version"`
	ReqInfo *ReqInfo `json:"requestInfo"`
	ReqItem *ReqItem `json:"requestItem"`
}

// ReqItem erp request item
type ReqItem struct {
	Method string      `json:"method"`
	Data   interface{} `json:"data"`
}

// ReturnReqData create device return order request item data
type ReturnReqData struct {
	DeptId    int      `json:"deptId"`
	AssetList []string `json:"assetList"`
	// IsEmergent emergent return or not
	// 1: emergent
	// 0: not emergent
	IsEmergent int `json:"isEmer"`
	// SkipConfirm skip double check or not
	// 1: skip double check
	// 0: do not skip double check
	SkipConfirm int    `json:"isPassDoubleCheck"`
	Reason      string `json:"reason"`
	// resonMsg is not typo, it's defined by erp
	ReasonMsg string `json:"resonMsg"`
	Remark    string `json:"remark"`
}

// OrderQueryReq erp order query request item data
type OrderQueryReqData struct {
	OrderId string   `json:"orderId"`
	AssetId []string `json:"assetId,omitempty"`
}
