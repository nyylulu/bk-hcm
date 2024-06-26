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
	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/tools/slice"
)

func (cli *client) getVpcMap(kt *kit.Kit, accountID string, region string,
	cloudVpcIDs []string) (map[string]*common.VpcDB, error) {

	vpcMap := make(map[string]*common.VpcDB)

	elems := slice.Split(cloudVpcIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		vpcParams := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  parts,
		}
		vpcFromDB, err := cli.listVpcFromDB(kt, vpcParams)
		if err != nil {
			return vpcMap, err
		}

		for _, vpc := range vpcFromDB {
			vpcMap[vpc.CloudID] = &common.VpcDB{
				VpcID:     vpc.ID,
				BkCloudID: vpc.BkCloudID,
			}
		}
	}

	return vpcMap, nil
}

func (cli *client) getSubnetMap(kt *kit.Kit, accountID string, region string,
	cloudSubnetsIDs []string) (map[string]string, error) {

	subnetMap := make(map[string]string)

	elems := slice.Split(cloudSubnetsIDs, constant.CloudResourceSyncMaxLimit)
	for _, parts := range elems {
		subnetParams := &SyncBaseParams{
			AccountID: accountID,
			Region:    region,
			CloudIDs:  parts,
		}
		subnetFromDB, err := cli.listSubnetFromDB(kt, subnetParams)
		if err != nil {
			return subnetMap, err
		}

		for _, subnet := range subnetFromDB {
			subnetMap[subnet.CloudID] = subnet.ID
		}
	}

	return subnetMap, nil
}
