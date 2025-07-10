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

package rollingserver

import (
	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablers "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
)

// NewRsAppliedRecord new rolling server applied record.
func NewRsAppliedRecord(dao dao.Set) *RsAppliedRecord {
	return &RsAppliedRecord{
		dao: dao,
	}
}

// RsAppliedRecord define rolling server applied record audit.
type RsAppliedRecord struct {
	dao dao.Set
}

// RsAppliedRecordUpdateAuditBuild rolling server applied record update audit build.
func (r *RsAppliedRecord) RsAppliedRecordUpdateAuditBuild(kt *kit.Kit, updates []protoaudit.CloudResourceUpdateInfo) (
	[]*tableaudit.AuditTable, error) {

	ids := make([]string, 0, len(updates))
	for _, one := range updates {
		ids = append(ids, one.ResID)
	}
	idMap, err := ListRsAppliedRecord(kt, r.dao, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		info, exist := idMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:    one.ResID,
			ResType:  enumor.RsAppliedRecordAuditResType,
			Action:   enumor.Update,
			BkBizID:  info.BkBizID,
			Vendor:   enumor.Ziyan,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data:    info,
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil
}

// ListRsAppliedRecord list rolling server applied record.
func ListRsAppliedRecord(kt *kit.Kit, dao dao.Set, ids []string) (map[string]tablers.RollingAppliedRecord, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ids),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.RollingAppliedRecord().List(kt, opt)
	if err != nil {
		logs.Errorf("list rolling server applied record failed, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablers.RollingAppliedRecord, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = converter.PtrToVal(one)
	}

	return result, nil
}
