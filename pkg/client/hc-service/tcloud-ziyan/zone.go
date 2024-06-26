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

package hcziyancli

import (
	"hcm/pkg/api/hc-service/zone"
	"hcm/pkg/client/common"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// NewZoneClient create a new zone api client.
func NewZoneClient(client rest.ClientInterface) *ZoneClient {
	return &ZoneClient{
		client: client,
	}
}

// ZoneClient is hc service zone api client.
type ZoneClient struct {
	client rest.ClientInterface
}

// SyncZone sync zone.
func (cli *ZoneClient) SyncZone(kt *kit.Kit, request *zone.TCloudZoneSyncReq) error {
	return common.RequestNoResp[zone.TCloudZoneSyncReq](cli.client, rest.POST, kt, request, "/zones/sync")
}
