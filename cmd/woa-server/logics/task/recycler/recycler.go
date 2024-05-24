/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package recycler ...
package recycler

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/language"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/classifier"
	"hcm/cmd/woa-server/logics/task/recycler/detector"
	"hcm/cmd/woa-server/logics/task/recycler/dispatcher"
	"hcm/cmd/woa-server/logics/task/recycler/returner"
	"hcm/cmd/woa-server/logics/task/recycler/transit"
	"hcm/cmd/woa-server/thirdparty"
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/cmd/woa-server/thirdparty/iamapi"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Interface recycler interface
type Interface interface {
	// RecycleCheck check whether hosts can be recycled or not
	RecycleCheck(kit *kit.Kit, param *types.RecycleCheckReq) (*types.RecycleCheckRst, error)
	// PreviewRecycleOrder preview resource recycle order
	PreviewRecycleOrder(kit *kit.Kit, param *types.PreviewRecycleReq) (*types.PreviewRecycleOrderRst, error)
	// AuditRecycleOrder audit resource recycle orders
	AuditRecycleOrder(kit *kit.Kit, param *types.AuditRecycleReq) error
	// CreateRecycleOrder create resource recycle order
	CreateRecycleOrder(kit *kit.Kit, param *types.CreateRecycleReq) (*types.CreateRecycleOrderRst, error)
	// GetRecycleOrder gets resource recycle order info
	GetRecycleOrder(kit *kit.Kit, param *types.GetRecycleOrderReq) (*types.GetRecycleOrderRst, error)
	// GetRecycleDetect gets resource recycle detection task info
	GetRecycleDetect(kit *kit.Kit, param *types.GetRecycleDetectReq) (*types.GetDetectTaskRst, error)
	// ListDetectHost gets recycle detection host list
	ListDetectHost(kit *kit.Kit, param *types.GetRecycleDetectReq) (*types.ListDetectHostRst, error)
	// GetRecycleDetectStep gets resource recycle detection step info
	GetRecycleDetectStep(kit *kit.Kit, param *types.GetDetectStepReq) (*types.GetDetectStepRst, error)

	// StartRecycleOrder starts resource recycle order
	StartRecycleOrder(kit *kit.Kit, param *types.StartRecycleOrderReq) error
	// StartDetectTask starts resource detection task
	StartDetectTask(kit *kit.Kit, param *types.StartDetectTaskReq) error
	// ReviseRecycleOrder revise recycle orders to remove detection failed hosts
	ReviseRecycleOrder(kit *kit.Kit, param *types.ReviseRecycleOrderReq) error
	// PauseRecycleOrder pauses resource recycle order
	PauseRecycleOrder(kit *kit.Kit, param mapstr.MapStr) error
	// ResumeRecycleOrder resumes resource recycle order
	ResumeRecycleOrder(kit *kit.Kit, param *types.ResumeRecycleOrderReq) error
	// TerminateRecycleOrder terminates resource recycle order
	TerminateRecycleOrder(kit *kit.Kit, param *types.TerminateRecycleOrderReq) error

	// GetRecycleHost gets resource recycle host info
	GetRecycleHost(kit *kit.Kit, param *types.GetRecycleHostReq) (*types.GetRecycleHostRst, error)
	// GetRecycleRecordDeviceType gets resource recycle record device type list
	GetRecycleRecordDeviceType(kit *kit.Kit) (*types.GetRecycleRecordDevTypeRst, error)
	// GetRecycleRecordRegion gets resource recycle record region list
	GetRecycleRecordRegion(kit *kit.Kit) (*types.GetRecycleRecordRegionRst, error)
	// GetRecycleRecordZone gets resource recycle record zone list
	GetRecycleRecordZone(kit *kit.Kit) (*types.GetRecycleRecordZoneRst, error)

	// GetRecycleBizHost gets business hosts in recycle module
	GetRecycleBizHost(kit *kit.Kit, param *types.GetRecycleBizHostReq) (*types.GetRecycleBizHostRst, error)

	// GetDetectStepCfg gets resource recycle step config info
	GetDetectStepCfg(kit *kit.Kit) (*types.GetDetectStepCfgRst, error)
}

// recycler provides resource recycle service
type recycler struct {
	lang language.CCLanguageIf
	//clientSet apimachinery.ClientSetInterface
	iam iamapi.IAMClientInterface
	cc  cmdb.Client
	cvm cvmapi.CVMClientInterface

	dispatcher *dispatcher.Dispatcher
}

// New create a recycler
func New(ctx context.Context, thirdCli *thirdparty.Client, esbCli esb.Client) (*recycler, error) {
	// new detector
	moduleDetector, err := detector.New(ctx, thirdCli, esbCli)
	if err != nil {
		return nil, err
	}

	// new returner
	moduleReturner, err := returner.New(ctx, thirdCli, esbCli)
	if err != nil {
		return nil, err
	}

	// new transit
	moduleTransit, err := transit.New(ctx, thirdCli, esbCli)

	// new dispatcher
	dispatch, err := dispatcher.New(ctx)
	if err != nil {
		return nil, err
	}
	dispatch.SetDetector(moduleDetector)
	dispatch.SetReturner(moduleReturner)
	dispatch.SetTransit(moduleTransit)

	recycler := &recycler{
		lang:       language.NewFromCtx(language.EmptyLanguageSetting),
		iam:        thirdCli.IAM,
		cc:         esbCli.Cmdb(),
		cvm:        thirdCli.CVM,
		dispatcher: dispatch,
	}

	return recycler, nil
}

// RecycleCheck check whether hosts can be recycled or not
func (r *recycler) RecycleCheck(kit *kit.Kit, param *types.RecycleCheckReq) (*types.RecycleCheckRst, error) {
	if kit.User == "" {
		logs.Errorf("failed to recycle check, for invalid user is empty, rid: %s", kit.Rid)
		return nil, errors.New("failed to recycle check, for invalid user is empty")
	}

	// 1. get host base info
	hostBase, err := r.getHostBaseInfo(param.IPs, param.AssetIDs, param.HostIDs)
	if err != nil {
		logs.Errorf("failed to recycle check, for list host err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	if len(hostBase) == 0 {
		return &types.RecycleCheckRst{Count: 0}, nil
	}

	hostIds := make([]int64, 0)
	for _, host := range hostBase {
		hostIds = append(hostIds, host.BkHostId)
	}

	// 2. get host topo info
	relations, err := r.getHostTopoInfo(hostIds)
	if err != nil {
		logs.Errorf("failed to recycle check, for list host err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	bizIds := make([]int64, 0)
	mapBizToModule := make(map[int64][]int64)
	mapHostToRel := make(map[int64]*cmdb.HostBizRel)
	for _, rel := range relations {
		mapHostToRel[rel.BkHostId] = rel
		if _, ok := mapBizToModule[rel.BkBizId]; !ok {
			mapBizToModule[rel.BkBizId] = []int64{rel.BkModuleId}
			bizIds = append(bizIds, rel.BkBizId)
		} else {
			mapBizToModule[rel.BkBizId] = append(mapBizToModule[rel.BkBizId], rel.BkModuleId)
		}
	}

	bizList, err := r.getBizInfo(bizIds)
	if err != nil {
		logs.Errorf("failed to recycle check, for get business info err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	mapBizIdToBiz := make(map[int64]*cmdb.BizInfo)
	for _, biz := range bizList {
		mapBizIdToBiz[biz.BkBizId] = biz
	}

	mapModuleIdToModule := make(map[int64]*cmdb.ModuleInfo)
	for bizId, moduleIds := range mapBizToModule {
		moduleIdUniq := util.IntArrayUnique(moduleIds)
		moduleList, err := r.getModuleInfo(kit, bizId, moduleIdUniq)
		if err != nil {
			logs.Errorf("failed to recycle check, for get module info err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		for _, module := range moduleList {
			mapModuleIdToModule[module.BkModuleId] = module
		}
	}

	// 3. check recycle permissions
	mapBizPermission := make(map[int64]bool)
	for _, bizId := range bizIds {
		hasPermission, err := r.hasRecyclePermission(kit, bizId)
		if err != nil {
			logs.Warnf("failed to check recycle permission, err: %v", err)
			continue
		}
		mapBizPermission[bizId] = hasPermission
	}

	// 4. check recyclability and create check result
	checkInfos := make([]*types.RecycleCheckInfo, 0)
	for _, host := range hostBase {
		bizId := int64(0)
		moduleId := int64(0)
		if rel, ok := mapHostToRel[host.BkHostId]; ok {
			bizId = rel.BkBizId
			moduleId = rel.BkModuleId
		}
		bizName := ""
		if biz, ok := mapBizIdToBiz[bizId]; ok {
			bizName = biz.BkBizName
		} else if bizId == 5000008 {
			// 业务资源池
			bizName = "业务资源池"
		}
		moduleName := ""
		if module, ok := mapModuleIdToModule[moduleId]; ok {
			moduleName = module.BkModuleName
		}
		hasPermission := false
		if permission, ok := mapBizPermission[bizId]; ok {
			hasPermission = permission
		}
		checkInfo := &types.RecycleCheckInfo{
			HostID:      host.BkHostId,
			AssetID:     host.BkAssetId,
			IP:          host.GetUniqIp(),
			BizID:       bizId,
			BizName:     bizName,
			TopoModule:  moduleName,
			Operator:    host.Operator,
			BakOperator: host.BakOperator,
			DeviceType:  host.SvrDeviceClass,
			State:       host.SvrStatus,
			InputTime:   host.SvrInputTime,
		}
		r.fillCheckInfo(checkInfo, kit.User, hasPermission)
		checkInfos = append(checkInfos, checkInfo)
	}

	rst := &types.RecycleCheckRst{
		Count: int64(len(checkInfos)),
		Info:  checkInfos,
	}

	return rst, nil
}

// fillCheckInfo fill host with recyclability check info
func (r *recycler) fillCheckInfo(host *types.RecycleCheckInfo, user string, hasPermission bool) {
	if !hasPermission {
		host.Recyclable = false
		host.Message = "无该业务资源回收权限，请至权限中心申请"
	} else if classifier.IsUnsupportedDevice(host.AssetID, host.IP) {
		host.Recyclable = false
		host.Message = "非YUNTI/SCR平台申请设备，请至具体申领平台回收"
	} else if host.TopoModule != "待回收" && host.TopoModule != "待回收模块" {
		host.Recyclable = false
		host.Message = "主机模块不是[待回收]"
	} else if strings.Contains(host.Operator, user) == false && strings.Contains(host.BakOperator, user) == false {
		host.Recyclable = false
		host.Message = "必须为主机负责人或备份负责人"
	} else {
		host.Recyclable = true
		host.Message = "可回收"
	}
}

// getHostBaseInfo get host detail info for recycle
func (r *recycler) getHostDetailInfo(ips, assetIds []string, hostIds []int64) ([]*table.RecycleHost,
	error) {

	// 1. get host base info
	hostBase, err := r.getHostBaseInfo(ips, assetIds, hostIds)
	if err != nil {
		logs.Errorf("failed to get host detail info, for list host err: %v", err)
		return nil, err
	}

	if len(hostBase) == 0 {
		return make([]*table.RecycleHost, 0), nil
	}

	bkHostIds := make([]int64, 0)
	for _, host := range hostBase {
		bkHostIds = append(bkHostIds, host.BkHostId)
	}

	// 2. get host biz info
	relations, err := r.getHostTopoInfo(bkHostIds)
	if err != nil {
		logs.Errorf("failed to get host detail info, for list host err: %v", err)
		return nil, err
	}

	bizIds := make([]int64, 0)
	mapHostToRel := make(map[int64]*cmdb.HostBizRel)
	for _, rel := range relations {
		bizIds = append(bizIds, rel.BkBizId)
		mapHostToRel[rel.BkHostId] = rel
	}
	bizIds = util.IntArrayUnique(bizIds)

	bizList, err := r.getBizInfo(bizIds)
	if err != nil {
		logs.Errorf("failed to get host detail info, for get business info err: %v", err)
		return nil, err
	}

	mapBizIdToBiz := make(map[int64]*cmdb.BizInfo)
	for _, biz := range bizList {
		mapBizIdToBiz[biz.BkBizId] = biz
	}

	// 3. fill host info
	hostDetails := make([]*table.RecycleHost, 0)
	cvmHosts := make([]*table.RecycleHost, 0)
	for _, host := range hostBase {
		bizId := int64(0)
		if rel, ok := mapHostToRel[host.BkHostId]; ok {
			bizId = rel.BkBizId
		}
		bizName := ""
		if biz, ok := mapBizIdToBiz[bizId]; ok {
			bizName = biz.BkBizName
		} else if bizId == 5000008 {
			// 业务资源池
			bizName = "业务资源池"
		}

		hostDetail := &table.RecycleHost{
			BizID:       bizId,
			BizName:     bizName,
			HostID:      host.BkHostId,
			AssetID:     host.BkAssetId,
			IP:          host.GetUniqIp(),
			DeviceType:  host.SvrDeviceClass,
			Zone:        host.BkZoneName,
			SubZone:     host.SubZone,
			ModuleName:  host.ModuleName,
			Operator:    host.Operator,
			BakOperator: host.BakOperator,
			InputTime:   host.SvrInputTime,
		}

		hostDetails = append(hostDetails, hostDetail)

		if classifier.IsQcloudCvm(hostDetail.AssetID) {
			cvmHosts = append(cvmHosts, hostDetail)
		}
	}

	// fill cvm info
	if err := r.fillCvmInfo(cvmHosts); err != nil {
		logs.Errorf("failed to fill cvm info, err: %v", err)
		return nil, err
	}

	return hostDetails, nil
}

// getHostBaseInfo get host base info in cc 3.0
func (r *recycler) getHostBaseInfo(ips, assetIds []string, hostIds []int64) ([]*cmdb.HostInfo, error) {
	rule := querybuilder.CombinedRule{
		Condition: querybuilder.ConditionOr,
		Rules:     make([]querybuilder.Rule, 0),
	}
	if len(ips) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.CombinedRule{
			Condition: querybuilder.ConditionAnd,
			Rules: []querybuilder.Rule{
				querybuilder.AtomRule{
					Field:    "bk_host_innerip",
					Operator: querybuilder.OperatorIn,
					Value:    ips,
				},
				// support bk_cloud_id 0 only
				querybuilder.AtomRule{
					Field:    "bk_cloud_id",
					Operator: querybuilder.OperatorEqual,
					Value:    0,
				},
			},
		})
	}
	if len(assetIds) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_asset_id",
			Operator: querybuilder.OperatorIn,
			Value:    assetIds,
		})
	}
	if len(hostIds) > 0 {
		rule.Rules = append(rule.Rules, querybuilder.AtomRule{
			Field:    "bk_host_id",
			Operator: querybuilder.OperatorIn,
			Value:    hostIds,
		})
	}

	req := &cmdb.ListHostReq{
		HostPropertyFilter: &querybuilder.QueryFilter{
			Rule: rule,
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			// 机型
			"svr_device_class",
			// 逻辑区域
			"logic_domain",
			"bk_zone_name",
			"sub_zone",
			"module_name",
			"raid_name",
			"svr_input_time",
			"operator",
			"bk_bak_operator",
			"srv_status",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: common.BKMaxInstanceLimit,
		},
	}

	resp, err := r.cc.ListHost(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc host info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc host info, err: %s", resp.ErrMsg)
	}

	return resp.Data.Info, nil
}

// getHostTopoInfo get host topo info in cc 3.0
func (r *recycler) getHostTopoInfo(hostIds []int64) ([]*cmdb.HostBizRel, error) {
	req := &cmdb.HostBizRelReq{
		BkHostId: hostIds,
	}

	resp, err := r.cc.FindHostBizRelation(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc host topo info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc host topo info, code: %d, msg: %s, rid: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc host topo info, err: %s", resp.ErrMsg)
	}

	return resp.Data, nil
}

// getBizInfo get business info in cc 3.0
func (r *recycler) getBizInfo(bizIds []int64) ([]*cmdb.BizInfo, error) {
	req := &cmdb.SearchBizReq{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_biz_id",
						Operator: querybuilder.OperatorIn,
						Value:    bizIds,
					},
				},
			},
		},
		Fields: []string{"bk_biz_id", "bk_biz_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 200,
		},
	}

	resp, err := r.cc.SearchBiz(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc business info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc business info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc business info, err: %s", resp.ErrMsg)
	}

	return resp.Data.Info, nil
}

// getModuleInfo get module info in cc 3.0
func (r *recycler) getModuleInfo(kit *kit.Kit, bizId int64, moduleIds []int64) ([]*cmdb.ModuleInfo, error) {
	req := &cmdb.SearchModuleReq{
		BkBizId: bizId,
		Condition: mapstr.MapStr{
			"bk_module_id": mapstr.MapStr{
				common.BKDBIN: moduleIds,
			},
		},
		Fields: []string{"bk_module_id", "bk_module_name"},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: 200,
		},
	}

	resp, err := r.cc.SearchModule(kit.Ctx, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc module info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc module info, code: %d, msg: %s, rid: %s", resp.Code, resp.ErrMsg, kit.Rid)
		return nil, fmt.Errorf("failed to get cc module info, err: %s", resp.ErrMsg)
	}

	return resp.Data.Info, nil
}

// PreviewRecycleOrder preview resource recycle order
func (r *recycler) PreviewRecycleOrder(kit *kit.Kit, param *types.PreviewRecycleReq) (*types.PreviewRecycleOrderRst,
	error) {

	// 1. get hosts info
	hosts, err := r.getHostDetailInfo(param.IPs, param.AssetIDs, param.HostIDs)
	if err != nil {
		logs.Errorf("failed to preview recycle order, for list host err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	// 2. classify hosts into groups with different recycle strategies
	groups, err := classifier.ClassifyRecycleGroups(hosts, param.ReturnPlan)
	if err != nil {
		logs.Errorf("failed to preview recycle order, for classify hosts err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	// 3. init and save recycle orders
	orders, err := r.initAndSaveRecycleOrders(kit, param.SkipConfirm, param.Remark, groups)
	if err != nil {
		logs.Errorf("failed to preview recycle order, init and save orders err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.PreviewRecycleOrderRst{
		Info: orders,
	}

	return rst, nil
}

// initRecycleOrder init and save recycle orders
func (r *recycler) initAndSaveRecycleOrders(kit *kit.Kit, skipConfirm bool, remark string,
	bizGroups map[int64]classifier.RecycleGroup) (

	[]*table.RecycleOrder, error) {

	now := time.Now()
	orders := make([]*table.RecycleOrder, 0)
	for biz, groups := range bizGroups {
		id, err := dao.Set().RecycleOrder().NextSequence(kit.Ctx)
		if err != nil {
			return nil, errf.New(common.CCErrObjectDBOpErrno, err.Error())
		}

		index := 1
		for grpType, group := range groups {
			if len(group) <= 0 {
				continue
			}

			// 1. init recycle order
			bizName := group[0].BizName
			order := &table.RecycleOrder{
				OrderID:       id,
				SuborderID:    fmt.Sprintf("%d-%d", id, index),
				BizID:         biz,
				BizName:       bizName,
				User:          kit.User,
				ResourceType:  classifier.MapGroupProperty[grpType].ResourceType,
				RecycleType:   classifier.MapGroupProperty[grpType].RecycleType,
				ReturnPlan:    classifier.MapGroupProperty[grpType].ReturnType,
				CostConcerned: classifier.MapGroupProperty[grpType].CostConcerned,
				SkipConfirm:   skipConfirm,
				Stage:         table.RecycleStageCommit,
				Status:        table.RecycleStatusUncommit,
				Handler:       "AUTO",
				TotalNum:      uint(len(group)),
				SuccessNum:    0,
				PendingNum:    0,
				FailedNum:     0,
				Remark:        remark,
				CreateAt:      now,
				UpdateAt:      now,
			}

			// 2. create and save recycle hosts
			if err := r.initAndSaveHosts(order, group); err != nil {
				logs.Errorf("failed to create recycle order for save recycle host err: %v, rid: %s", err, kit.Rid)
				return nil, fmt.Errorf("failed to create recycle order for save recycle host err: %v", err)
			}

			// 3. save recycle order
			if err := dao.Set().RecycleOrder().CreateRecycleOrder(kit.Ctx, order); err != nil {
				logs.Errorf("failed to create recycle order for save recycle order err: %v, rid: %s", err, kit.Rid)
				return nil, fmt.Errorf("failed to create recycle order for save recycle order err: %v", err)
			}

			orders = append(orders, order)
			index++
		}
	}

	return orders, nil
}

// initAndSaveHosts inits and saves recycle hosts
func (r *recycler) initAndSaveHosts(order *table.RecycleOrder, hosts []*table.RecycleHost) error {
	now := time.Now()
	costRate := 0.0
	if order.ResourceType == table.ResourceTypePm && order.RecycleType == table.RecycleTypeRegular {
		costRate = 0.6
	}
	for _, host := range hosts {
		host.OrderID = order.OrderID
		host.SuborderID = order.SuborderID
		host.User = order.User
		host.Stage = order.Stage
		host.Status = order.Status
		host.ReturnCostRate = costRate
		host.CreateAt = now
		host.UpdateAt = now

		if err := dao.Set().RecycleHost().CreateRecycleHost(context.Background(), host); err != nil {
			logs.Errorf("failed to save recycle host, ip: %s, err: %v", host.IP, err)
			return fmt.Errorf("failed to save recycle host, ip: %s, err: %v", host.IP, err)
		}
	}

	return nil
}

func (r *recycler) fillCvmInfo(hosts []*table.RecycleHost) error {
	// skip query cvm instance when input no hosts
	if len(hosts) == 0 {
		return nil
	}

	ipList := make([]string, 0)
	mapIp2Host := make(map[string]*table.RecycleHost)
	for _, host := range hosts {
		ipList = append(ipList, host.IP)
		mapIp2Host[host.IP] = host
	}

	req := &cvmapi.InstanceQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmInstanceStatusMethod,
		},
		Params: &cvmapi.InstanceQueryParam{
			LanIp: ipList,
		},
	}

	resp, err := r.cvm.QueryCvmInstances(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to query cvm instance, err: %v", err)
		return err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("query cvm failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
		return fmt.Errorf("query cvm failed, code: %d, msg: %s", resp.Error.Code, resp.Error.Message)
	}

	for _, inst := range resp.Result.Data {
		if host, ok := mapIp2Host[inst.LanIp]; ok {
			host.InstID = inst.InstanceId
			host.ObsProject = inst.ObsProject
			if inst.Pool == 1 {
				host.Pool = table.PoolPublic
			} else {
				host.Pool = table.PoolPrivate
			}
		}
	}

	return nil
}

// AuditRecycleOrder audit resource recycle orders
func (r *recycler) AuditRecycleOrder(kit *kit.Kit, param *types.AuditRecycleReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			common.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)

	if cnt != 1 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("get invalid recycle order count %d != 1", cnt)
	}

	// verify order status
	order := insts[0]
	if order.Status != table.RecycleStatusAudit {
		logs.Errorf("failed to audit recycle order, for order status %s is not %s", order.Status,
			table.RecycleStatusAudit)
		return fmt.Errorf("failed to audit recycle order, for order status %s is not %s", order.Status,
			table.RecycleStatusAudit)
	}

	// verify operator permission
	operator := kit.User
	if operator != "dommyzhang" && operator != "forestchen" {
		logs.Errorf("failed to audit recycle order, for operator has no permission")
		return errors.New("failed to audit recycle order, for operator has no permission")
	}

	task := dispatcher.NewTask(order.Status)
	taskCtx := &dispatcher.AuditContext{
		Order:      order,
		Dispatcher: r.dispatcher,
		Operator:   operator,
		Approval:   param.Approval,
		Remark:     param.Remark,
	}
	if err := task.State.Execute(taskCtx); err != nil {
		logs.Errorf("failed to execute audit task, err: %v, order id: %s, state: %s", err, order.SuborderID,
			task.State.Name())
		return err
	}

	return nil
}

// CreateRecycleOrder create resource recycle order
func (r *recycler) CreateRecycleOrder(kit *kit.Kit, param *types.CreateRecycleReq) (*types.CreateRecycleOrderRst,
	error) {

	// 1. get hosts info
	hosts, err := r.getHostDetailInfo(param.IPs, param.AssetIDs, param.HostIDs)
	if err != nil {
		logs.Errorf("failed to create recycle order, for list host err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	if len(hosts) == 0 {
		logs.Errorf("get no valid host to create recycle order")
		return nil, errors.New("get no valid host to create recycle order")
	}

	// 2. check permission
	bizIds := make([]int64, 0)
	for _, host := range hosts {
		bizIds = append(bizIds, host.BizID)
	}
	bizIds = util.IntArrayUnique(bizIds)

	for _, bizId := range bizIds {
		hasPermission, err := r.hasRecyclePermission(kit, bizId)
		if err != nil {
			logs.Errorf("failed to check recycle permission, err: %v", err)
			return nil, err
		}

		if !hasPermission {
			logs.Errorf("has no permission to recycle resource in biz %d", bizId)
			return nil, fmt.Errorf("has no permission to recycle resource in biz %d", bizId)
		}
	}

	// 3. classify hosts into groups with different recycle strategies
	groups, err := classifier.ClassifyRecycleGroups(hosts, param.ReturnPlan)
	if err != nil {
		logs.Errorf("failed to preview recycle order, for classify hosts err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	// 4. init and save recycle orders
	orders, err := r.initAndSaveRecycleOrders(kit, param.SkipConfirm, param.Remark, groups)
	if err != nil {
		logs.Errorf("failed to preview recycle order, init and save orders err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	r.setOrderCommitted(orders)

	rst := &types.CreateRecycleOrderRst{
		Info: orders,
	}

	return rst, nil
}

// GetRecycleOrder gets resource recycle order info
func (r *recycler) GetRecycleOrder(kit *kit.Kit, param *types.GetRecycleOrderReq) (*types.GetRecycleOrderRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle order, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetRecycleOrderRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().RecycleOrder().CountRecycleOrder(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recycle order count, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.RecycleOrder, 0)
		return rst, nil
	}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetRecycleDetect gets resource recycle detection task info
func (r *recycler) GetRecycleDetect(kit *kit.Kit, param *types.GetRecycleDetectReq) (*types.GetDetectTaskRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle detection task, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetDetectTaskRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().DetectTask().CountDetectTask(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recycle detection task count, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.DetectTask, 0)
		return rst, nil
	}

	insts, err := dao.Set().DetectTask().FindManyDetectTask(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle detection task, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// ListDetectHost gets recycle detection host list
func (r *recycler) ListDetectHost(kit *kit.Kit, param *types.GetRecycleDetectReq) (*types.ListDetectHostRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle detection task, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	insts, err := dao.Set().DetectTask().GetRecycleHostList(kit.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get recycle detection task, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.ListDetectHostRst{
		Info: insts,
	}

	return rst, nil
}

// GetRecycleDetectStep gets resource recycle detection step info
func (r *recycler) GetRecycleDetectStep(kit *kit.Kit, param *types.GetDetectStepReq) (*types.GetDetectStepRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle detection step, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetDetectStepRst{}
	if param.Page.EnableCount {
		cnt, err := dao.Set().DetectStep().CountDetectStep(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recycle detection step count, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.DetectStep, 0)
		return rst, nil
	}

	insts, err := dao.Set().DetectStep().FindManyDetectStep(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle detection step, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// StartRecycleOrder starts resource recycle order
func (r *recycler) StartRecycleOrder(kit *kit.Kit, param *types.StartRecycleOrderReq) error {
	filter := map[string]interface{}{}
	if len(param.OrderID) > 0 {
		filter["order_id"] = mapstr.MapStr{
			common.BKDBIN: param.OrderID,
		}
	}

	if len(param.SuborderID) > 0 {
		filter["suborder_id"] = mapstr.MapStr{
			common.BKDBIN: param.SuborderID,
		}
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no recycle order to start")
	}

	r.setOrderNextStatus(insts)

	return nil
}

func (r *recycler) setOrderNextStatus(orders []*table.RecycleOrder) {
	now := time.Now()
	for _, order := range orders {
		nextStatus := order.Status
		failedNum := uint(0)
		switch order.Status {
		case table.RecycleStatusUncommit:
			nextStatus = table.RecycleStatusCommitted
		case table.RecycleStatusDetectFailed:
			nextStatus = table.RecycleStatusDetecting
		case table.RecycleStatusTransitFailed:
			nextStatus = table.RecycleStatusTransiting
		case table.RecycleStatusReturnFailed:
			nextStatus = table.RecycleStatusReturning
			failedNum = order.FailedNum
		default:
			logs.Warnf("failed to set order %s to next status, for unsupported status %s", order.SuborderID,
				order.Status)
			continue
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"failed_num": failedNum,
			"status":     nextStatus,
			"update_at":  now,
		}

		// do not dispatch order to start if set next status failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s committed, err: %v", order.SuborderID, err)
			continue
		}

		// add order to dispatch queue
		r.dispatcher.Add(order.SuborderID)
	}
}

func (r *recycler) setOrderCommitted(orders []*table.RecycleOrder) {
	now := time.Now()
	for _, order := range orders {
		// need not set order committed if it's not uncommit
		if order.Status != table.RecycleStatusUncommit {
			logs.Warnf("failed to set order %s committed, for invalid status %s", order.SuborderID, order.Status)
			continue
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"status":    table.RecycleStatusCommitted,
			"update_at": now,
		}

		// do not dispatch order to start if set committed failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s committed, err: %v", order.SuborderID, err)
			continue
		}

		// add order to dispatch queue
		r.dispatcher.Add(order.SuborderID)
	}
}

// StartDetectTask starts resource detection task
func (r *recycler) StartDetectTask(kit *kit.Kit, param *types.StartDetectTaskReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			common.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no recycle order to start")
	}

	// check status
	for _, order := range insts {
		// cannot restart detection task if it's not detect failed
		if order.Status != table.RecycleStatusDetectFailed {
			logs.Errorf("cannot restart order %s detection task, for its status %s not %s", order.SuborderID,
				order.Status, table.RecycleStatusDetectFailed)
			return fmt.Errorf("cannot restart order %s detection task, for its status %s not %s", order.SuborderID,
				order.Status, table.RecycleStatusDetectFailed)
		}
	}

	// set order status detecting
	r.setOrderDetecting(insts)

	return nil
}

func (r *recycler) setOrderDetecting(orders []*table.RecycleOrder) {
	now := time.Now()
	for _, order := range orders {
		// need not set order detecting if it's not detect failed
		if order.Status != table.RecycleStatusDetectFailed {
			logs.Warnf("failed to set order %s committed, for invalid status %s", order.SuborderID, order.Status)
			continue
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"failed_num": 0,
			"status":     table.RecycleStatusDetecting,
			"update_at":  now,
		}

		// do not dispatch order to start if set committed failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s detecting, err: %v", order.SuborderID, err)
			continue
		}

		// add order to dispatch queue
		r.dispatcher.Add(order.SuborderID)
	}
}

// ReviseRecycleOrder revise recycle orders to remove detection failed hosts
func (r *recycler) ReviseRecycleOrder(kit *kit.Kit, param *types.ReviseRecycleOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			common.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no recycle order to start")
	}

	// check status
	for _, order := range insts {
		// cannot restart detection task if it's not detect failed
		if order.Status != table.RecycleStatusDetectFailed {
			logs.Errorf("cannot restart order %s detection task, for its status %s not %s", order.SuborderID,
				order.Status, table.RecycleStatusDetectFailed)
			return fmt.Errorf("cannot restart order %s detection task, for its status %s not %s", order.SuborderID,
				order.Status, table.RecycleStatusDetectFailed)
		}
	}

	// remove detection failed hosts and set order status detecting
	if err := r.reviseOrder(insts); err != nil {
		logs.Errorf("failed to revise recycle order, err: %v", err)
		return fmt.Errorf("failed to revise recycle order, err: %v", err)
	}

	return nil
}

func (r *recycler) reviseOrder(orders []*table.RecycleOrder) error {
	now := time.Now()
	for _, order := range orders {
		// need not set order detecting if it's not detect failed
		if order.Status != table.RecycleStatusDetectFailed {
			logs.Warnf("failed to set order %s committed, for invalid status %s", order.SuborderID, order.Status)
			continue
		}

		// get detection failed ips
		filter := map[string]interface{}{
			"suborder_id": order.SuborderID,
			"status":      table.DetectStatusFailed,
		}
		ips, err := dao.Set().DetectTask().GetRecycleHostList(context.Background(), filter)
		if err != nil {
			logs.Errorf("failed to get order %s detection failed host list, err: %v", order.SuborderID, err)
			return fmt.Errorf("failed to get order %s detection failed host list, err: %v", order.SuborderID, err)
		}

		failedNum := uint(len(ips))
		if failedNum >= order.TotalNum {
			logs.Errorf("cannot revise order %s, for all detection task failed", order.SuborderID)
			return fmt.Errorf("cannot revise order %s, for all detection task failed", order.SuborderID)
		}

		// remove detection failed hosts
		if err := r.removeDetectFailedHosts(order.SuborderID, ips); err != nil {
			logs.Errorf("failed to remove detection failed hosts, err: %v", err)
			return fmt.Errorf("failed to remove detection failed hosts, err: %v", err)
		}

		leftNum := order.TotalNum - failedNum
		orderFilter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"total_num":   leftNum,
			"success_num": 0,
			"pending_num": leftNum,
			"failed_num":  0,
			"status":      table.RecycleStatusDetecting,
			"update_at":   now,
		}

		// do not dispatch order to start if set committed failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), orderFilter, update); err != nil {
			logs.Errorf("failed to set order %s detecting, err: %v", order.SuborderID, err)
			return fmt.Errorf("failed to set order detecting, err: %v", err)
		}

		// add order to dispatch queue
		r.dispatcher.Add(order.SuborderID)
	}

	return nil
}

func (r *recycler) removeDetectFailedHosts(orderID string, ips []interface{}) error {
	filter := map[string]interface{}{
		"suborder_id": orderID,
		"ip": map[string]interface{}{
			common.BKDBIN: ips,
		},
	}

	if _, err := dao.Set().RecycleHost().DeleteRecycleHost(context.Background(), filter); err != nil {
		logs.Errorf("failed to delete recycle host, err: %v", err)
		return err
	}

	if _, err := dao.Set().DetectTask().DeleteDetectTask(context.Background(), filter); err != nil {
		logs.Errorf("failed to delete detection task, err: %v", err)
		return err
	}

	if _, err := dao.Set().DetectStep().DeleteDetectStep(context.Background(), filter); err != nil {
		logs.Errorf("failed to delete detection task step, err: %v", err)
		return err
	}

	return nil
}

func (r *recycler) getRecycleTaskById(kit *kit.Kit, taskId string) (*table.DetectTask, error) {
	filter := map[string]interface{}{
		"task_id": taskId,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: 1,
	}

	insts, err := dao.Set().DetectTask().FindManyDetectTask(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle task, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	cnt := len(insts)
	if cnt != 1 {
		logs.Errorf("get invalid recycle task count %d != 1, rid: %s", cnt, kit.Rid)
		return nil, fmt.Errorf("get invalid recycle task count %d != 1, rid: %s", cnt, kit.Rid)
	}

	return insts[0], nil
}

// PauseRecycleOrder pauses resource recycle order
func (r *recycler) PauseRecycleOrder(kit *kit.Kit, param mapstr.MapStr) error {
	// TODO
	return nil
}

// ResumeRecycleOrder resumes resource recycle order
func (r *recycler) ResumeRecycleOrder(kit *kit.Kit, param *types.ResumeRecycleOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			common.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no recycle order to terminate")
	}

	// add order to dispatch queue
	for _, order := range insts {
		r.dispatcher.Add(order.SuborderID)
	}

	return nil
}

// TerminateRecycleOrder terminates resource recycle order
func (r *recycler) TerminateRecycleOrder(kit *kit.Kit, param *types.TerminateRecycleOrderReq) error {
	filter := map[string]interface{}{
		"suborder_id": mapstr.MapStr{
			common.BKDBIN: param.SuborderID,
		},
	}

	page := metadata.BasePage{}

	insts, err := dao.Set().RecycleOrder().FindManyRecycleOrder(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle order, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	cnt := len(insts)
	if cnt == 0 {
		logs.Errorf("get invalid recycle order count %d != 1, rid: %s", cnt, kit.Rid)
		return fmt.Errorf("found no recycle order to terminate")
	}

	// check status
	for _, order := range insts {
		// cannot terminate detection task if it's not detect failed
		switch order.Status {
		case table.RecycleStatusDone, table.RecycleStatusTerminate:
			logs.Errorf("need not terminate order %s, for its status %s", order.SuborderID, order.Status)
			return fmt.Errorf("need not terminate order %s, for its status %s", order.SuborderID, order.Status)
		case table.RecycleStatusTransiting, table.RecycleStatusReturning:
			logs.Errorf("cannot terminate order %s, for its status %s", order.SuborderID, order.Status)
			return fmt.Errorf("cannot terminate order %s, for its status %s", order.SuborderID, order.Status)
		}
	}

	// set order status terminate
	if err := r.terminateOrder(insts); err != nil {
		logs.Errorf("failed to revise recycle order, err: %v", err)
		return fmt.Errorf("failed to revise recycle order, err: %v", err)
	}

	return nil
}

func (r *recycler) terminateOrder(orders []*table.RecycleOrder) error {
	now := time.Now()
	for _, order := range orders {
		switch order.Status {
		case table.RecycleStatusDone, table.RecycleStatusTerminate:
			logs.Errorf("need not terminate order %s, for its status %s", order.SuborderID, order.Status)
			return fmt.Errorf("need not terminate order %s, for its status %s", order.SuborderID, order.Status)
		case table.RecycleStatusTransiting, table.RecycleStatusReturning:
			logs.Errorf("cannot terminate order %s, for its status %s", order.SuborderID, order.Status)
			return fmt.Errorf("cannot terminate order %s, for its status %s", order.SuborderID, order.Status)
		}

		filter := &mapstr.MapStr{
			"suborder_id": order.SuborderID,
		}

		update := &mapstr.MapStr{
			"stage":     table.RecycleStageTerminate,
			"status":    table.RecycleStatusTerminate,
			"update_at": now,
		}

		// do not dispatch order to start if set committed failed
		if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), filter, update); err != nil {
			logs.Warnf("failed to set order %s detecting, err: %v", order.SuborderID, err)
			return fmt.Errorf("failed to terminate order %s, err:%v", order.SuborderID, err)
		}
	}

	return nil
}

// GetRecycleHost gets resource recycle host info
func (r *recycler) GetRecycleHost(kit *kit.Kit, param *types.GetRecycleHostReq) (*types.GetRecycleHostRst, error) {
	filter, err := param.GetFilter()
	if err != nil {
		logs.Errorf("failed to get recycle host, for get filter err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetRecycleHostRst{}

	if param.Page.EnableCount {
		cnt, err := dao.Set().RecycleHost().CountRecycleHost(kit.Ctx, filter)
		if err != nil {
			logs.Errorf("failed to get recycle host count, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		rst.Count = int64(cnt)
		rst.Info = make([]*table.RecycleHost, 0)
		return rst, nil
	}

	insts, err := dao.Set().RecycleHost().FindManyRecycleHost(kit.Ctx, param.Page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle host, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}
	rst.Count = 0
	rst.Info = insts

	return rst, nil
}

// GetRecycleRecordDeviceType gets resource recycle record device type list
func (r *recycler) GetRecycleRecordDeviceType(kit *kit.Kit) (*types.GetRecycleRecordDevTypeRst, error) {
	filter := map[string]interface{}{}
	insts, err := dao.Set().RecycleHost().Distinct(kit.Ctx, "device_type", filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetRecycleRecordDevTypeRst{
		Info: insts,
	}

	return rst, nil
}

// GetRecycleRecordRegion gets resource recycle record region list
func (r *recycler) GetRecycleRecordRegion(kit *kit.Kit) (*types.GetRecycleRecordRegionRst, error) {
	filter := map[string]interface{}{}
	insts, err := dao.Set().RecycleHost().Distinct(kit.Ctx, "bk_zone_name", filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetRecycleRecordRegionRst{
		Info: insts,
	}

	return rst, nil
}

// GetRecycleRecordZone gets resource recycle record zone list
func (r *recycler) GetRecycleRecordZone(kit *kit.Kit) (*types.GetRecycleRecordZoneRst, error) {
	filter := map[string]interface{}{}
	insts, err := dao.Set().RecycleHost().Distinct(kit.Ctx, "sub_zone", filter)
	if err != nil {
		return nil, err
	}

	rst := &types.GetRecycleRecordZoneRst{
		Info: insts,
	}

	return rst, nil
}

// GetRecycleBizHost gets business hosts in recycle module
func (r *recycler) GetRecycleBizHost(kit *kit.Kit, param *types.GetRecycleBizHostReq) (*types.GetRecycleBizHostRst,
	error) {

	req := &cmdb.ListBizHostReq{
		BkBizId: param.BizID,
		ModuleCond: []cmdb.ConditionItem{
			// recycle module's default is 3
			{
				Field:    "default",
				Operator: "$eq",
				Value:    3,
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"operator",
			"bk_bak_operator",
			"svr_device_class",
			"sub_zone",
			"svr_input_time",
			"srv_status",
		},
		Page: cmdb.BasePage{
			Start: param.Page.Start,
			Limit: param.Page.Limit,
		},
	}

	if param.Page.EnableCount {
		req.Page.Start = 0
		req.Page.Limit = 1
	}

	resp, err := r.cc.ListBizHost(kit.Ctx, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc host info, code: %d, msg: %s, rid: %s", resp.Code, resp.ErrMsg, kit.Rid)
		return nil, fmt.Errorf("failed to get cc host info, err: %s", resp.ErrMsg)
	}

	hostList := make([]*types.RecycleBizHost, 0)
	rst := new(types.GetRecycleBizHostRst)
	if param.Page.EnableCount {
		rst.Count = int64(resp.Data.Count)
		return rst, nil
	}

	for _, ccHost := range resp.Data.Info {
		host := &types.RecycleBizHost{
			HostID:      ccHost.BkHostId,
			AssetID:     ccHost.BkAssetId,
			IP:          ccHost.GetUniqIp(),
			Operator:    ccHost.Operator,
			BakOperator: ccHost.BakOperator,
			DeviceType:  ccHost.SvrDeviceClass,
			SubZone:     ccHost.SubZone,
			State:       ccHost.SvrStatus,
			InputTime:   ccHost.SvrInputTime,
		}
		hostList = append(hostList, host)
	}

	rst.Info = hostList
	return rst, nil
}

// GetDetectStepCfg gets resource recycle detection step config info
func (r *recycler) GetDetectStepCfg(kit *kit.Kit) (*types.GetDetectStepCfgRst, error) {
	filter := map[string]interface{}{
		"enable": true,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: common.BKNoLimit,
	}

	insts, err := dao.Set().DetectStepCfg().GetDetectStepConfig(kit.Ctx, page, filter)
	if err != nil {
		logs.Errorf("failed to get recycle detection step, err: %v, rid: %s", err, kit.Rid)
		return nil, err
	}

	rst := &types.GetDetectStepCfgRst{
		Info: insts,
	}

	return rst, nil
}

func (r *recycler) hasRecyclePermission(kit *kit.Kit, bizId int64) (bool, error) {
	user := kit.User
	if user == "" {
		logs.Errorf("failed to check permission, for invalid user is empty, rid: %s", kit.Rid)
		return false, errors.New("failed to check permission, for invalid user is empty")
	}

	req := &iamapi.AuthVerifyReq{
		System: "bk_cr",
		Subject: &iamapi.Subject{
			Type: "user",
			ID:   user,
		},
		Action: &iamapi.Action{
			ID: "resource_recycle",
		},
		Resources: []*iamapi.Resource{
			&iamapi.Resource{
				System: "bk_cmdb",
				Type:   "biz",
				ID:     strconv.Itoa(int(bizId)),
			},
		},
	}
	resp, err := r.iam.AuthVerify(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to auth verify, err: %v, rid: %s", err, kit.Rid)
		return false, err
	}
	if resp.Code != 0 {
		logs.Errorf("failed to auth verify, code: %d, msg: %s, rid: %s", resp.Code, resp.Message, kit.Rid)
		return false, fmt.Errorf("failed to auth verify, err: %s", resp.Message)
	}

	if resp.Data.Allowed != true {
		return false, nil
	}

	return true, nil
}
