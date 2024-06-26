// Copyright (c) 2017-2018 THL A29 Limited, a Tencent company. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v20181217

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

const APIVersion = "2018-12-17"

// Client BPaas client
type Client struct {
	common.Client
}

// NewClient New BPaas client
func NewClient(credential common.CredentialIface, region string, clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

// NewGetBpaasApplicationDetailRequest ...
func NewGetBpaasApplicationDetailRequest() (request *GetBpaasApplicationDetailRequest) {
	request = &GetBpaasApplicationDetailRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}

	request.Init().WithApiInfo("bpaas", APIVersion, "GetBpaasApplicationDetail")

	return
}

// NewGetBpaasApplicationDetailResponse ...
func NewGetBpaasApplicationDetailResponse() (response *GetBpaasApplicationDetailResponse) {
	response = &GetBpaasApplicationDetailResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return

}

// GetBpaasApplicationDetail 查看申请详情
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR_ACCOUNTERROR = "InternalError.AccountError"
//	INTERNALERROR_CAUTHERROR = "InternalError.CauthError"
//	INVALIDPARAMETER_IDNOTEXIST = "InvalidParameter.IdNotExist"
//	UNAUTHORIZEDOPERATION_PERMISSIONDENIED = "UnauthorizedOperation.PermissionDenied"
func (c *Client) GetBpaasApplicationDetail(request *GetBpaasApplicationDetailRequest) (response *GetBpaasApplicationDetailResponse, err error) {
	return c.GetBpaasApplicationDetailWithContext(context.Background(), request)
}

// GetBpaasApplicationDetailWithContext  查看申请详情
//
// 可能返回的错误码:
//
//	FAILEDOPERATION = "FailedOperation"
//	INTERNALERROR_ACCOUNTERROR = "InternalError.AccountError"
//	INTERNALERROR_CAUTHERROR = "InternalError.CauthError"
//	INVALIDPARAMETER_IDNOTEXIST = "InvalidParameter.IdNotExist"
//	UNAUTHORIZEDOPERATION_PERMISSIONDENIED = "UnauthorizedOperation.PermissionDenied"
func (c *Client) GetBpaasApplicationDetailWithContext(ctx context.Context, request *GetBpaasApplicationDetailRequest) (response *GetBpaasApplicationDetailResponse, err error) {
	if request == nil {
		request = NewGetBpaasApplicationDetailRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("GetBpaasApplicationDetail require credential")
	}

	request.SetContext(ctx)

	response = NewGetBpaasApplicationDetailResponse()
	err = c.Send(request, response)
	return
}
