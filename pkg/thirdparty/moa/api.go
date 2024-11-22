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

package moa

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Request 发起验证
func (m *moaCli) Request(kt *kit.Kit, req *InitiateVerificationReq) (*InitiateVerificationResp, error) {
	resp, err := call[InitiateVerificationResp](m.client, m.config, rest.POST, kt, req, nil, "/request")
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Verify 验证结果
func (m *moaCli) Verify(kt *kit.Kit, req *VerificationReq) (*VerificationResp, error) {
	resp, err := call[VerificationResp](m.client, m.config, rest.POST, kt, req, nil, "/verify")
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func call[OT any](cli rest.ClientInterface, cfg *cc.MOA, method rest.VerbType, kt *kit.Kit, req any,
	params map[string]string, url string, urlParams ...any) (*OT, error) {

	header := getCommonHeader(kt, cfg)
	resp := new(MOAResp[*OT])
	err := cli.Verb(method).
		SubResourcef(url, urlParams...).
		WithContext(kt.Ctx).
		WithHeaders(header).
		WithParams(params).
		Body(req).
		Do().Into(resp)

	if err != nil {
		logs.Errorf("fail to call api gateway api, err: %v, url: %s, rid: %s", err, url, kt.Rid)
		return nil, err
	}

	if resp.ErrCode != 0 {
		err := fmt.Errorf("failed to call moa api , code: %d, msg: %s, response: %v",
			resp.ErrCode, resp.ErrMsg, resp)
		logs.Errorf("api gateway returns error, url: %s, err: %v, rid: %s", url, err, kt.Rid)
		return nil, err
	}
	return resp.Data, nil
}

func getCommonHeader(kt *kit.Kit, cfg *cc.MOA) http.Header {
	header := kt.Header()

	timestamp := fmt.Sprintf("%d", time.Now().Unix()) // 生成时间戳，注意服务器的时间与标准时间差不能大于180秒
	r := rand.New(rand.NewSource(time.Now().Unix()))
	nonce := strconv.Itoa(r.Intn(4096)) // 随机字符串，十分钟内不重复即可
	signStr := fmt.Sprintf("%s%s%s%s", timestamp, cfg.Token, nonce, timestamp)
	sign := fmt.Sprintf("%X", sha256.Sum256([]byte(signStr))) // 输出大写的结果

	header.Add("x-rio-paasid", cfg.PaasID)
	header.Add("x-rio-nonce", nonce)
	header.Add("x-rio-timestamp", timestamp)
	header.Add("x-rio-signature", sign)

	return header
}

// MOAResp response struct for moa api
type MOAResp[OT any] struct {
	ErrCode int    `json:"ErrCode"`
	ErrMsg  string `json:"ErrMsg"`
	Data    OT     `json:"Data"`
}
