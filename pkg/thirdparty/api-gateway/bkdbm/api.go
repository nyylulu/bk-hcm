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

// Package bkdbm ...
package bkdbm

import (
	"strconv"
	"strings"

	"hcm/pkg/kit"
	"hcm/pkg/rest"
	apigateway "hcm/pkg/thirdparty/api-gateway"
)

// QueryMachinePool query machine pool.
// @doc https://bkapigw.woa.com/docs/api-docs/gateway/bkdbm?apiName=query_machine_pool
func (c *dbm) QueryMachinePool(kt *kit.Kit, req *ListMachinePool) (*ListMachinePoolResp, error) {
	var hostIDStrs []string
	for _, hostID := range req.HostIDs {
		hostIDStrs = append(hostIDStrs, strconv.FormatInt(hostID, 10))
	}
	hostIDJoinStr := strings.Join(hostIDStrs, ",")

	params := map[string]string{
		"bk_host_ids": hostIDJoinStr,
		"ips":         strings.Join(req.IPs, ","),
		"offset":      strconv.FormatInt(req.Offset, 10),
		"limit":       strconv.FormatInt(req.Limit, 10),
	}
	return apigateway.ApiGatewayCallWithoutReq[ListMachinePoolResp](c.client, c.config, rest.GET,
		kt, params, "/db_dirty/query_machine_pool")
}
