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

package tag

import (
	"errors"
	"fmt"
	"time"

	"hcm/cmd/hc-service/logics/res-sync/ziyan"
	typecore "hcm/pkg/adaptor/types/core"
	typessg "hcm/pkg/adaptor/types/security-group"
	typestag "hcm/pkg/adaptor/types/tag"
	"hcm/pkg/api/core"
	apitag "hcm/pkg/api/hc-service/tag"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

type resSyncGroupByKey struct {
	Region  string
	ResType enumor.CloudResourceType
}

// TCloudZiyanBatchTagRes 给账号下多个资源打多个标签
func (t *tag) TCloudZiyanBatchTagRes(cts *rest.Contexts) (interface{}, error) {
	req := new(apitag.TCloudBatchTagResRequest)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	account, err := t.cs.DataService().TCloudZiyan.Account.Get(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("fail to get ziyan account info: %s, err: %v, rid: %s", req.AccountID, err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	if account.Vendor != enumor.TCloudZiyan {
		return nil, errf.Newf(errf.InvalidParameter, "account %s is not tcloud-ziyan account", req.AccountID)
	}

	resourceGroupByMap := make(map[resSyncGroupByKey][]string)

	resourceList := make([]string, len(req.Resources))
	for i, resTmp := range req.Resources {
		resourceList[i] = req.Resources[i].Convert(account.Extension.CloudMainAccountID)

		resKey := resSyncGroupByKey{
			Region:  resTmp.Region,
			ResType: resTmp.ResType,
		}
		resourceGroupByMap[resKey] = append(resourceGroupByMap[resKey], resTmp.ResCloudID)
	}
	client, err := t.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("fail to get tcloud-ziyan adaptor: %v, rid: %s", err, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	opt := &typestag.TCloudTagResOpt{
		ResourceList: resourceList,
		Tags:         req.Tags,
	}
	resp, err := client.TagResources(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	// 将云上最新的标签版本同步到本地
	err = t.syncResourceTag(cts.Kit, resourceGroupByMap, req.AccountID, req.Tags)
	if err != nil {
		logs.Errorf("failed to sync resources after tagging resources, err: %v, res: %+v, rid: %s", err,
			req.Resources, cts.Kit.Rid)
		return nil, err
	}

	return resp, nil
}

// syncResourceTag 将云上最新的标签同步回本地
func (t *tag) syncResourceTag(kt *kit.Kit, resGroupByMap map[resSyncGroupByKey][]string, accountID string,
	tags []core.TagPair) error {

	// 打标签后，同步一次，将云上标签同步回本地
	syncCli, err := t.syncCli.TCloudZiyan(kt, accountID)
	if err != nil {
		logs.Warnf("failed to sync resources after tagging resources, err: %v, res: %+v, rid: %s", err,
			resGroupByMap, kt.Rid)
		return errors.New("failed to sync resources after tagging resources")
	}

	for key, resources := range resGroupByMap {
		params := &ziyan.SyncBaseParams{
			AccountID: accountID,
			Region:    key.Region,
			CloudIDs:  resources,
		}

		switch key.ResType {
		case enumor.SecurityGroupCloudResType:
			if err := t.syncSecurityGroupTag(kt, syncCli, params, core.NewTagMap(tags...)); err != nil {
				logs.Warnf("failed to sync security group after tagging resources, err: %v, "+
					"res cloud_ids: %v, rid: %s", err, resources, kt.Rid)
			}
		// TODO 其他资源打标签时也需要触发一次同步
		default:
			// 为了提醒后续资源打标签时在这里补充逻辑，这里返回error
			logs.Errorf("unsupported sync resource type %s after tagging resources, res cloud_ids: %v, rid: %s",
				key.ResType, resources, kt.Rid)
			return fmt.Errorf("unsupported sync resource type %s after tagging resources", key.ResType)
		}
	}

	return nil
}

func (t *tag) syncSecurityGroupTag(kt *kit.Kit, syncCli ziyan.Interface, syncParams *ziyan.SyncBaseParams,
	tagsMap core.TagMap) error {

	if err := syncParams.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	ziyanCli, err := t.ad.TCloudZiyan(kt, syncParams.AccountID)
	if err != nil {
		logs.Errorf("fail to get tcloud-ziyan adaptor: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	listOpt := &typessg.TCloudListOption{
		Region:   syncParams.Region,
		CloudIDs: syncParams.CloudIDs,
		Page: &typecore.TCloudPage{
			Limit: typecore.TCloudQueryLimit,
		},
	}

	// 云上标签更新有延迟，轮询云上资源，直到云上标签版本与本地一致
	startWaiting := time.Now()
	isMatch := false
outside:
	for time.Now().Before(startWaiting.Add(constant.IntervalWaitResourceSync)) {
		time.Sleep(time.Millisecond * 300)

		cloudSGs, err := ziyanCli.ListSecurityGroupNew(kt, listOpt)
		if err != nil {
			logs.Errorf("failed to list security group from tcloud-ziyan, err: %v, res_ids: %v, rid: %s", err,
				listOpt.CloudIDs, kt.Rid)
			return err
		}

		for _, cloudSG := range cloudSGs {
			matchTag := 0
			for _, cloudTag := range cloudSG.TagSet {
				value, ok := tagsMap[cvt.PtrToVal(cloudTag.Key)]
				if !ok {
					continue
				}
				matchTag += 1
				if value != cvt.PtrToVal(cloudTag.Value) {
					continue outside
				}
			}
			if matchTag != len(tagsMap) {
				continue outside
			}
		}

		isMatch = true
		break
	}

	if !isMatch {
		logs.Errorf("failed to wait tcloud-ziyan security group sync: timeout, res cloud_ids: %v, rid: %s",
			syncParams.CloudIDs, kt.Rid)
		return errors.New("failed to wait tcloud-ziyan security group sync: timeout")
	}

	// 同步安全组资源到本地
	if _, err := syncCli.SecurityGroup(kt, syncParams, new(ziyan.SyncSGOption)); err != nil {
		logs.Errorf("failed to sync sg after tagging resources, err: %v, res cloud_ids: %v, rid: %s", err,
			syncParams.CloudIDs, kt.Rid)
		return err
	}

	return nil
}
