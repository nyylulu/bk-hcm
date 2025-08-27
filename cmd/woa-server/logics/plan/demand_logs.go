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

package plan

import (
	"fmt"
	"strconv"

	ptypes "hcm/cmd/woa-server/types/plan"
	tasktypes "hcm/cmd/woa-server/types/task"
	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rpdc "hcm/pkg/dal/table/resource-plan/res-plan-demand-changelog"
	tableTypes "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/times"

	"github.com/jmoiron/sqlx"
)

// ListCrpDemandChangeLog list crp demand change log by demand id
func (c *Controller) ListCrpDemandChangeLog(kt *kit.Kit, req *ptypes.ListDemandChangeLogReq) (
	*ptypes.ListDemandChangeLogResp, error) {

	listRules := make([]*filter.AtomRule, 0)
	listRules = append(listRules, tools.RuleEqual("demand_id", req.DemandID))

	listReq := &rpproto.DemandChangelogListReq{
		ListReq: core.ListReq{
			Filter: tools.ExpressionAnd(listRules...),
			Page:   req.Page,
		},
	}

	rst, err := c.client.DataService().Global.ResourcePlan.ListDemandChangelog(kt, listReq)
	if err != nil {
		logs.Errorf("failed to list crp demand change log, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &ptypes.ListDemandChangeLogResp{Count: rst.Count}, nil
	}

	rstDetails := make([]*ptypes.ListDemandChangeLogItem, len(rst.Details))
	for idx, tableDetail := range rst.Details {
		expectTimeStr, err := times.TransTimeStrWithLayout(strconv.Itoa(tableDetail.ExpectTime),
			constant.DateLayoutCompact, constant.DateLayout)
		if err != nil {
			logs.Errorf("failed to convert expect time to string, err: %v, expect time: %d, rid: %s", err,
				tableDetail.ExpectTime, kt.Rid)
			return nil, err
		}

		rstDetails[idx] = &ptypes.ListDemandChangeLogItem{
			ID:                tableDetail.ID,
			DemandId:          tableDetail.DemandID,
			ExpectTime:        expectTimeStr,
			ObsProject:        tableDetail.ObsProject,
			RegionName:        tableDetail.RegionName,
			ZoneName:          tableDetail.ZoneName,
			DeviceType:        tableDetail.DeviceType,
			ChangeCvmAmount:   tableDetail.OSChange.Decimal,
			ChangeCoreAmount:  cvt.PtrToVal(tableDetail.CpuCoreChange),
			ChangeRamAmount:   cvt.PtrToVal(tableDetail.MemoryChange),
			ChangedDiskAmount: cvt.PtrToVal(tableDetail.DiskSizeChange),
			DemandSource:      tableDetail.Type.Name(),
			TicketID:          tableDetail.TicketID,
			CrpSn:             tableDetail.CrpOrderID,
			SuborderID:        tableDetail.SuborderID,
			CreateTime:        tableDetail.CreatedAt.String(),
			Remark:            tableDetail.Remark,
		}
	}

	return &ptypes.ListDemandChangeLogResp{
		Count:   0,
		Details: rstDetails,
	}, nil
}

// AddMatchedPlanDemandExpendLogs add matched plan demand expend logs.
func (c *Controller) AddMatchedPlanDemandExpendLogs(kt *kit.Kit, bkBizID int64, subOrder *tasktypes.ApplyOrder,
	verifyGroups []VerifyResPlanElemV2) error {
	// if resource type is not cvm,	return success.
	if subOrder.ResourceType != tasktypes.ResourceTypeCvm &&
		subOrder.ResourceType != tasktypes.ResourceTypeUpgradeCvm {
		return nil
	}

	verifySuborder := tasktypes.Suborder{
		ResourceType:   subOrder.ResourceType,
		Spec:           subOrder.Spec,
		UpgradeCVMList: subOrder.UpgradeCVMList,
	}

	verifySlice, err := c.fillVerifyElems(kt, verifySuborder, bkBizID, subOrder.ObsProject, verifyGroups)
	if err != nil {
		logs.Errorf("failed to fill verify elems, err: %v, bkBizID: %d, subOrder: %+v, rid: %s", err, bkBizID,
			cvt.PtrToVal(subOrder), kt.Rid)
		return err
	}

	// call verify resource plan demands to verify each cvm demands.
	ret, err := c.VerifyProdDemandsV2(kt, bkBizID, subOrder.RequireType, verifySlice)
	if err != nil {
		logs.Errorf("failed to get matched resource plan demand ids, err: %v, bkBizID: %d, subOrder: %+v, rid: %s",
			err, bkBizID, cvt.PtrToVal(subOrder), kt.Rid)
		return err
	}

	if len(ret) != len(verifySlice) {
		return errf.Newf(errf.InvalidParameter, "get matched plan demand result length: %d is "+
			"not eq verifySlice: %d, verify result: %+v", len(ret), len(verifySlice), ret)
	}

	for i, verifyElem := range verifySlice {
		demandIDs := ret[i].MatchDemandIDs
		if len(demandIDs) == 0 {
			logs.Infof("get matched plan demand ids empty, bkBizID: %d, demandIDs: %v, subOrder: %+v, rid: %s",
				bkBizID, demandIDs, cvt.PtrToVal(subOrder), kt.Rid)
			continue
		}

		// 记录日志方便排查问题
		logs.Infof("get matched plan demand ids success, bkBizID: %d, demandIDs: %v, verifyElem: %+v, subOrder: %+v, "+
			"rid: %s", bkBizID, demandIDs, verifyElem, cvt.PtrToVal(subOrder), kt.Rid)

		err = c.addMatchedPlanDemandExpendLogs(kt, demandIDs, bkBizID, subOrder, verifyElem)
		if err != nil {
			logs.Errorf("failed to add matched plan demand expend logs, err: %v, bkBizID: %d, demandIDs: %v, "+
				"verifyElem: %+v, order id: %s, rid: %s", err, bkBizID, demandIDs, verifyElem, subOrder.SubOrderId,
				kt.Rid)
			return err
		}
	}
	return nil
}

func (c *Controller) addMatchedPlanDemandExpendLogs(kt *kit.Kit, demandIDs []string, bkBizID int64,
	subOrder *tasktypes.ApplyOrder, verifyElem VerifyResPlanElemV2) error {

	demadOpt := &types.ListOption{
		Filter: tools.ContainersExpression("id", demandIDs),
		Page:   core.NewDefaultBasePage(),
	}
	demandListResp, err := c.dao.ResPlanDemand().List(kt, demadOpt)
	if err != nil {
		logs.Errorf("list resource plan demand by ids failed, err: %v, demandIDs: %v, rid: %s", err, demandIDs, kt.Rid)
		return fmt.Errorf("list resource plan demand by ids failed, err: %v, demandIDs: %v", err, demandIDs)
	}
	if len(demandListResp.Details) == 0 {
		logs.Infof("matched list plan demand from db ids empty, bkBizID: %d, demandIDs: %v, subOrder: %+v, rid: %s",
			bkBizID, demandIDs, cvt.PtrToVal(subOrder), kt.Rid)
		return nil
	}

	inserts := make([]rpdc.DemandChangelogTable, len(demandIDs))
	for idx, demandItem := range demandListResp.Details {
		inserts[idx] = rpdc.DemandChangelogTable{
			DemandID:   demandItem.ID,
			SuborderID: subOrder.SubOrderId,
			Type:       enumor.DemandChangelogTypeExpend,
			ExpectTime: demandItem.ExpectTime,
			// TODO 暂时先用ID赋值，后续整合可以再查一次region表查出name
			ObsProject:     subOrder.ObsProject,
			RegionName:     verifyElem.RegionID,
			ZoneName:       verifyElem.ZoneID,
			DeviceType:     verifyElem.DeviceType,
			CpuCoreChange:  cvt.ValToPtr(-verifyElem.CpuCore),
			OSChange:       &tableTypes.Decimal{},
			MemoryChange:   cvt.ValToPtr(int64(0)),
			DiskSizeChange: cvt.ValToPtr(int64(0)),
			Remark:         subOrder.Remark,
		}
	}
	_, err = c.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		return c.dao.ResPlanDemandChangelog().CreateWithTx(kt, txn, inserts)
	})
	if err != nil {
		logs.Errorf("failed to create plan crp demand log, err: %v, bkBizID: %d, demandIDs: %v, subOrder: %+v, rid: %s",
			err, bkBizID, demandIDs, cvt.PtrToVal(subOrder), kt.Rid)
	}
	return nil
}
