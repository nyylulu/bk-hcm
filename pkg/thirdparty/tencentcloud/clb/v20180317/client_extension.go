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
	"context"
	"errors"

	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
)

// NewDescribeSlaCapacityRequest ...
func NewDescribeSlaCapacityRequest() (request *DescribeSlaCapacityRequest) {
	request = &DescribeSlaCapacityRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}

	request.Init().WithApiInfo("clb", APIVersion, "DescribeSlaCapacity")

	return
}

// NewDescribeSlaCapacityResponse ...
func NewDescribeSlaCapacityResponse() (response *DescribeSlaCapacityResponse) {
	response = &DescribeSlaCapacityResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return

}

// DescribeSlaCapacity
// 查询性能保障规格参数
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR = "InternalError"
//	INVALIDPARAMETER = "InvalidParameter"
//	INVALIDPARAMETER_FORMATERROR = "InvalidParameter.FormatError"
//	INVALIDPARAMETERVALUE = "InvalidParameterValue"
func (c *Client) DescribeSlaCapacityWithContext(ctx context.Context,
	request *DescribeSlaCapacityRequest) (response *DescribeSlaCapacityResponse, err error) {
	if request == nil {
		request = NewDescribeSlaCapacityRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeSlaCapacity require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeSlaCapacityResponse()
	err = c.Send(request, response)
	return
}
