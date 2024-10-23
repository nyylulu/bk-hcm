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

package itsm

import (
	"fmt"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway"
)

// TicketResult ticket result.
type TicketResult struct {
	SN            string `json:"sn"`
	TicketURL     string `json:"ticket_url"`
	CurrentStatus string `json:"current_status"`
	ApproveResult bool   `json:"approve_result"`
}

type queryTicketResp struct {
	apigateway.BaseResponse `json:",inline"`
	Data                    []TicketResult `json:"data"`
}

func (i *itsm) batchQueryTicketResult(kt *kit.Kit, sns []string) (results []TicketResult, err error) {
	req := map[string]interface{}{"sn": sns}
	resp := new(queryTicketResp)
	err = i.client.Post().
		SubResourcef("/ticket_approval_result/").
		WithContext(kt.Ctx).
		WithHeaders(i.header(kt)).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return results, err
	}
	if !resp.Result || resp.Code != 0 {
		return results, fmt.Errorf("query ticket result failed, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// GetTicketResults 批量查询单据结果
func (i *itsm) GetTicketResults(kt *kit.Kit, sns []string) (result []TicketResult, err error) {
	results, err := i.batchQueryTicketResult(kt, sns)
	if err != nil {
		logs.Errorf("failed to get itsm ticket results, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if len(results) == 0 {
		logs.Errorf("itsm returns empty result, err: %v, rid: %s", errf.RecordNotFound, kt.Rid)
		return result, err
	}
	return results, nil
}

// GetTicketResult 查询单据结果
func (i *itsm) GetTicketResult(kt *kit.Kit, sn string) (result TicketResult, err error) {
	results, err := i.batchQueryTicketResult(kt, []string{sn})
	if err != nil {
		return result, err
	}
	if len(results) == 0 {
		return result, errf.New(errf.RecordNotFound, "itsm returns empty result")
	}
	return results[0], nil
}

// GetTicketStatus get itsm ticket status
func (i *itsm) GetTicketStatus(kt *kit.Kit, sn string) (*GetTicketStatusResp, error) {
	resp := &struct {
		apigateway.BaseResponse `json:",inline"`
		Data                    *GetTicketStatusResp `json:"data"`
	}{}

	param := map[string]string{
		"sn": sn,
	}
	err := i.client.Get().
		SubResourcef("/get_ticket_status/").
		WithParams(param).
		WithContext(kt.Ctx).
		WithHeaders(i.header(kt)).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("get ticket status failed, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data, err
}
