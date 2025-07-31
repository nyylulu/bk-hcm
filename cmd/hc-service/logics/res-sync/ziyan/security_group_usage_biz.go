/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	"fmt"

	usagebizrelmgr "hcm/cmd/hc-service/logics/res-sync/usage-biz-rel"
	securitygroup "hcm/pkg/adaptor/types/security-group"
	"hcm/pkg/api/core"
	cloudcore "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
	"hcm/pkg/ziyan"
)

// SecurityGroupUsageBiz 通过安全组关联资源的业务ID，更新安全组使用业务ID. 安全组使用业务同步完毕后，会根据安全组的的标签,自动分配业务.
func (cli *client) SecurityGroupUsageBiz(kt *kit.Kit, params *SyncSGUsageBizParams) error {

	mgr := usagebizrelmgr.NewUsageBizRelManager(cli.dbCli)
	for i := range params.SGList {
		sg := &params.SGList[i]
		err := mgr.SyncSecurityGroupUsageBiz(kt, sg)
		if err != nil {
			logs.Errorf("sync security group usage biz failed, err: %v, sg: %+v, rid: %s", err, sg, kt.Rid)
			return err
		}
	}

	sgIDs := slice.Map(params.SGList, cloudcore.BaseSecurityGroup.GetID)
	sgList, err := cli.listSGByIDs(kt, sgIDs)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, sgIDs: %v, rid: %s", err, sgIDs, kt.Rid)
		return err
	}
	if err := cli.autoAssignSGToBiz(kt, sgList); err != nil {
		logs.Errorf("validate sg with mgmt info failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// SyncSGUsageBizParams 同步安全组使用业务参数，使用业务只依赖本地数据
type SyncSGUsageBizParams struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	SGList    []cloudcore.BaseSecurityGroup
}

func (cli *client) listSGByIDs(kt *kit.Kit, sgIDs []string) ([]cloudcore.BaseSecurityGroup, error) {
	listReq := &protocloud.SecurityGroupListReq{
		Filter: tools.ExpressionAnd(tools.RuleIn("id", sgIDs)),
		Page:   core.NewDefaultBasePage(),
	}
	sgResp, err := cli.dbCli.Global.SecurityGroup.ListSecurityGroup(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return sgResp.Details, nil
}

func (cli *client) updateSGMgmtAttr(kt *kit.Kit, sg *cloudcore.BaseSecurityGroup) error {
	updateReq := &protocloud.SecurityGroupBatchUpdateReq[cloudcore.TCloudSecurityGroupExtension]{
		SecurityGroups: []protocloud.SecurityGroupBatchUpdate[cloudcore.TCloudSecurityGroupExtension]{
			{
				ID:         sg.ID,
				MgmtType:   sg.MgmtType,
				MgmtBizID:  sg.MgmtBizID,
				BkBizID:    sg.BkBizID,
				Manager:    sg.Manager,
				BakManager: sg.BakManager,
			},
		},
	}
	if err := cli.dbCli.TCloudZiyan.SecurityGroup.BatchUpdateSecurityGroup(kt, updateReq); err != nil {
		logs.Errorf("[%s] request dataservice BatchUpdateSecurityGroup failed, err: %v, rid: %s", enumor.TCloudZiyan,
			err, kt.Rid)
		return err
	}
	return nil
}

// 根据管理类型等信息自动分配安全组到业务中
func (cli *client) autoAssignSGToBiz(kt *kit.Kit, sgFromDB []cloudcore.BaseSecurityGroup) error {

	for i := range sgFromDB {
		local := &sgFromDB[i]
		if isSkipAssignSG(local) {
			continue
		}

		isChange, err := cli.assignSGToBiz(kt, local)
		if err != nil {
			logs.Errorf("assign sg to biz failed, sgID: %s, err: %v, rid: %s", local.ID, err, kt.Rid)
			return err
		}
		if isChange {
			if err := cli.updateSGMgmtAttr(kt, local); err != nil {
				logs.Errorf("update security group mgmt biz id failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}
		}
	}
	return nil
}

func isSkipAssignSG(local *cloudcore.BaseSecurityGroup) bool {
	// 白名单校验，遇白直接跳过
	if slice.IsItemInSlice(cc.HCService().SecurityGroupSkipList, local.CloudID) {
		return true
	}
	if local.MgmtType == enumor.MgmtTypePlatform {
		// 本地标记为平台管理的安全组, 跳过自动分配逻辑
		return true
	}
	if local.MgmtType == enumor.MgmtTypeBiz && local.BkBizID != constant.UnassignedBiz {
		// 已分配的安全组, 跳过自动分配逻辑
		return true
	}
	return false
}

// assignSGToBiz 根据云上安全组信息，自动分配业务ID, 如果条件不符合没有发生触发自动分配, isChange=false
func (cli *client) assignSGToBiz(kt *kit.Kit, sg *cloudcore.BaseSecurityGroup) (isChange bool, err error) {
	meta := ziyan.ParseResourceMetaIgnoreErr(sg.Tags)
	err = meta.Validate()
	if err != nil {
		// 标签不完整, 不进行自动分配
		// 标签不完整, 无需向上抛error
		return false, nil
	}

	bizIds, err := ziyan.GetBkBizIdByBs2(kt, cli.cmdbCli, []int64{meta.Bs2NameID})
	if err != nil {
		logs.Errorf("fail to get bkBizId by bs2NameIds for clb, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}
	if len(bizIds) == 0 {
		return false, fmt.Errorf("bizIds not found")
	}

	// 安全组使用业务不唯一, 不自动分配业务
	if len(sg.UsageBizIDs) > 1 {
		// 绑定了多个业务资源
		return false, nil
	}

	enable, err := cli.validateSGAssociationStatistic(kt, sg)
	if err != nil {
		logs.Errorf("validate sg association statistic failed, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}
	if !enable {
		return false, nil
	}

	sg.BakManager = meta.BakManager
	sg.Manager = meta.Manager
	if len(sg.UsageBizIDs) == 0 || sg.UsageBizIDs[0] == bizIds[0] {
		sg.BkBizID = bizIds[0]
		sg.MgmtBizID = bizIds[0]
		sg.MgmtType = enumor.MgmtTypeBiz
	}
	return true, nil
}

// 检查安全组绑定的资源类型, 仅绑定了cvm、clb两种资源且数量和本地sgCommonRels的数量一致时才进行自动分配
func (cli *client) validateSGAssociationStatistic(kt *kit.Kit, sg *cloudcore.BaseSecurityGroup) (
	enable bool, err error) {

	opt := &securitygroup.TCloudListOption{
		Region:   sg.Region,
		CloudIDs: []string{sg.CloudID},
	}
	statistics, err := cli.cloudCli.DescribeSGAssociationStatistics(kt, opt)
	if err != nil {
		logs.Errorf("describe security group association statistics failed, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}
	if len(statistics) != 1 {
		logs.Errorf("security group association statistics length is not 1, len: %d, rid: %s", len(statistics), kt.Rid)
		return false, fmt.Errorf("security group association statistics length expect 1, but got %d", len(statistics))
	}
	sgStatistic := statistics[0]

	totalCount := converter.PtrToVal(sgStatistic.TotalCount)
	if totalCount != converter.PtrToVal(sgStatistic.CLB)+converter.PtrToVal(sgStatistic.CVM) {
		return false, nil
	} else if totalCount == 0 {
		return true, nil
	}

	req := &core.ListReq{
		Filter: tools.EqualExpression("security_group_id", sg.ID),
		Page:   core.NewCountPage(),
	}
	resp, err := cli.dbCli.Global.SGCommonRel.ListSgCommonRels(kt, req)
	if err != nil {
		logs.Errorf("list sg common rels failed, err: %v, rid: %s", err, kt.Rid)
		return false, err
	}

	if resp.Count == totalCount {
		return true, nil
	}

	return false, nil
}
