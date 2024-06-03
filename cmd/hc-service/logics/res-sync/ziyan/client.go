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

package ziyan

import (
	ziyan "hcm/pkg/adaptor/tcloud-ziyan"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/kit"
)

// Interface support resource sync.
type Interface interface {
	CloudCli() ziyan.TCloudZiyan

	SecurityGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGOption) (*SyncResult, error)
	RemoveSecurityGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	SecurityGroupRule(kt *kit.Kit, params *SyncBaseParams, opt *SyncSGRuleOption) (*SyncResult, error)

	Zone(kt *kit.Kit, opt *SyncZoneOption) (*SyncResult, error)

	Region(kt *kit.Kit, opt *SyncRegionOption) (*SyncResult, error)

	ArgsTplAddress(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error)
	RemoveArgsTplAddressDeleteFromCloud(kt *kit.Kit, accountID string, region string) error
	ArgsTplAddressGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error)
	RemoveArgsTplAddressGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error
	ArgsTplService(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error)
	RemoveArgsTplServiceDeleteFromCloud(kt *kit.Kit, accountID string, region string) error
	ArgsTplServiceGroup(kt *kit.Kit, params *SyncBaseParams, opt *SyncArgsTplOption) (*SyncResult, error)
	RemoveArgsTplServiceGroupDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	Cert(kt *kit.Kit, params *SyncBaseParams, opt *SyncCertOption) (*SyncResult, error)
	RemoveCertDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	LoadBalancer(kt *kit.Kit, params *SyncBaseParams, opt *SyncLBOption) (*SyncResult, error)
	RemoveLoadBalancerDeleteFromCloud(kt *kit.Kit, accountID string, region string) error

	LoadBalancerWithListener(kt *kit.Kit, params *SyncBaseParams, opt *SyncLBOption) (*SyncResult, error)
	Listener(kt *kit.Kit, opt *SyncListenerOfSingleLBOption) (*SyncResult, error)
}

var _ Interface = new(client)

type client struct {
	accountID string
	cloudCli  ziyan.TCloudZiyan
	dbCli     *dataservice.Client
}

// CloudCli return tcloud client.
func (cli *client) CloudCli() ziyan.TCloudZiyan {
	return cli.cloudCli
}

// NewClient new sync client.
func NewClient(dbCli *dataservice.Client, cloudCli ziyan.TCloudZiyan) Interface {
	return &client{
		dbCli:    dbCli,
		cloudCli: cloudCli,
	}
}
