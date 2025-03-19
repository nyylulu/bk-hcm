/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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

package v20180317

import (
	"encoding/json"

	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

// DescribeSlaCapacityRequest ...
type DescribeSlaCapacityRequest struct {
	*tchttp.BaseRequest

	SlaTypes []string `json:"SlaTypes,omitnil,omitempty" name:"SlaTypes"`
}

// ToJsonString ...
func (r *DescribeSlaCapacityRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeSlaCapacityRequest) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}

// SlaItem ...
type SlaItem struct {
	SlaType    string `json:"SlaType,omitnil,omitempty" name:"SlaType"`
	SlaName    string `json:"SlaName,omitnil,omitempty" name:"SlaName"`
	MaxConn    int64  `json:"MaxConn,omitnil,omitempty" name:"MaxConn"`
	MaxCps     int64  `json:"MaxCps,omitnil,omitempty" name:"MaxCps"`
	MaxOutBits int64  `json:"MaxOutBits,omitnil,omitempty" name:"MaxOutBits"`
	MaxInBits  int64  `json:"MaxInBits,omitnil,omitempty" name:"MaxInBits"`
	MaxQps     int64  `json:"MaxQps,omitnil,omitempty" name:"MaxQps"`
}

// DescribeSlaCapacityResponseParams ...
type DescribeSlaCapacityResponseParams struct {
	// 可用区支持的资源列表。
	SlaSet []*SlaItem `json:"SlaSet,omitnil,omitempty" name:"SlaSet"`

	// 唯一请求 ID，由服务端生成，每次请求都会返回（若请求因其他原因未能抵达服务端，则该次请求不会获得 RequestId）。
	// 定位问题时需要提供该次请求的 RequestId。
	RequestId *string `json:"RequestId,omitnil,omitempty" name:"RequestId"`
}

type DescribeSlaCapacityResponse struct {
	*tchttp.BaseResponse
	Response *DescribeSlaCapacityResponseParams `json:"Response"`
}

func (r *DescribeSlaCapacityResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

// FromJsonString It is highly **NOT** recommended to use this function
// because it has no param check, nor strict type check
func (r *DescribeSlaCapacityResponse) FromJsonString(s string) error {
	return json.Unmarshal([]byte(s), &r)
}
