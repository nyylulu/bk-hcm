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

package host

import (
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/dissolve/host"
	define "hcm/pkg/dal/table/dissolve/host"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/caiche"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// RecycledHost provides interface for operations of recycle host.
type RecycledHost interface {
	Create(kt *kit.Kit, hosts []define.RecycleHostTable) ([]string, error)
	Update(kt *kit.Kit, host *define.RecycleHostTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*host.ListRecycleHostDetails, error)
	Delete(kt *kit.Kit, ids []string) error
	IsDissolveHost(kt *kit.Kit, assetIDs []string) (map[string]bool, error)
	Sync(kt *kit.Kit) error
}

type logics struct {
	dao          dao.Set
	thirdCli     *thirdparty.Client
	projectNames []string
}

// New create recycle host logics.
func New(dao dao.Set, thirdCli *thirdparty.Client, projectNames []string) RecycledHost {
	return &logics{
		dao:          dao,
		thirdCli:     thirdCli,
		projectNames: projectNames,
	}
}

// Create recycle host.
func (l *logics) Create(kt *kit.Kit, hosts []define.RecycleHostTable) ([]string, error) {
	result, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		res, err := l.dao.RecycleHost().CreateWithTx(kt, txn, hosts)
		if err != nil {
			logs.Errorf("create recycle host failed, err: %v, data: %+v, rid: %s", err, hosts, kt.Rid)
			return nil, err
		}

		return res, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// Update recycle host.
func (l *logics) Update(kt *kit.Kit, host *define.RecycleHostTable) error {
	_, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		err := l.dao.RecycleHost().UpdateWithTx(kt, txn, tools.EqualExpression("id", host.ID), host)
		if err != nil {
			logs.Errorf("update recycle host failed, err: %v, data: %+v, rid: %s", err, host, kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}

// List recycle host.
func (l *logics) List(kt *kit.Kit, opt *types.ListOption) (
	*host.ListRecycleHostDetails, error) {

	return l.dao.RecycleHost().List(kt, opt)
}

// Delete recycle host.
func (l *logics) Delete(kt *kit.Kit, ids []string) error {
	_, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		err := l.dao.RecycleHost().DeleteWithTx(kt, txn, tools.ContainersExpression("id", ids))
		if err != nil {
			logs.Errorf("delete recycle host failed, err: %v, ids: %+v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}

// IsDissolveHost check if host is dissolve host.
func (l *logics) IsDissolveHost(kt *kit.Kit, assetIDs []string) (map[string]bool, error) {
	result := make(map[string]bool)

	for _, ids := range slice.Split(assetIDs, int(core.DefaultMaxPageLimit)) {
		req := &types.ListOption{
			Filter: tools.ContainersExpression("asset_id", ids),
			Fields: []string{"asset_id"},
			Page:   core.NewDefaultBasePage(),
		}

		list, err := l.List(kt, req)
		if err != nil {
			logs.Errorf("list recycle host failed, err: %v, ids: %+v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}

		for _, one := range list.Details {
			result[converter.PtrToVal(one.AssetID)] = true
		}
	}

	return result, nil
}

// Sync recycle host.
func (l *logics) Sync(kt *kit.Kit) error {
	start := time.Now()
	logs.Infof("start sync recycle host, time: %v, rid: %s", start, kt.Rid)

	dbHosts, err := l.getAllHostFromDB(kt)
	if err != nil {
		logs.Errorf("get recycle host from db failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	caiCheHosts, err := l.getAllHostFromCaiChe(kt)
	if err != nil {
		logs.Errorf("get recycle host from caiche failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	create, update, deleteIDs := diff(dbHosts, caiCheHosts)
	if len(deleteIDs) != 0 {
		for _, split := range slice.Split(deleteIDs, int(core.DefaultMaxPageLimit)) {
			if err = l.Delete(kt, split); err != nil {
				logs.Errorf("delete recycle host failed, err: %v, ids: %+v, rid: %s", err, split, kt.Rid)
				return err
			}
		}
	}

	for _, one := range update {
		if err = l.Update(kt, &one); err != nil {
			logs.Errorf("update recycle host failed, err: %v, data: %+v, rid: %s", err, one, kt.Rid)
			return err
		}
	}

	if len(create) != 0 {
		for _, split := range slice.Split(create, int(core.DefaultMaxPageLimit)) {
			if _, err = l.Create(kt, split); err != nil {
				logs.Errorf("create recycle host failed, err: %v, ids: %+v, rid: %s", err, split, kt.Rid)
				return err
			}
		}
	}

	end := time.Now()
	logs.Infof("end sync recycle host, time: %v, cost: %v, create: %d, update: %d, delete: %d, rid: %s", end,
		end.Sub(start), len(create), len(update), len(deleteIDs), kt.Rid)

	return nil
}

func (l *logics) getAllHostFromDB(kt *kit.Kit) ([]define.RecycleHostTable, error) {
	hosts := make([]define.RecycleHostTable, 0)
	req := &types.ListOption{Filter: tools.AllExpression(),
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "id"}}

	for {
		result, err := l.List(kt, req)
		if err != nil {
			logs.Errorf("get recycle host failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
			return nil, err
		}

		hosts = append(hosts, result.Details...)

		if len(result.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	return hosts, nil
}

func (l *logics) getAllHostFromCaiChe(kt *kit.Kit) ([]define.RecycleHostTable, error) {
	hosts := make([]define.RecycleHostTable, 0)

	req := &caiche.ListDeviceReq{
		VirtualDepartmentName: []string{"IEG_Global", "IEG技术运营部"},
		AbolishPhase: []enumor.AbolishPhase{enumor.Incomplete, enumor.Complete, enumor.BsiComplete,
			enumor.Retain},
		ProjectName: l.projectNames,
		PageIndex:   1, // 裁撤系统这个api是从1开始的
		PageSize:    core.DefaultMaxPageLimit,
	}

	for {
		resp, err := l.thirdCli.CaiChe.ListDevice(kt, req)
		if err != nil {
			logs.Errorf("get recycle host failed, err: %v, req: %+v, rid: %s", err, *req, kt.Rid)
			return nil, err
		}

		hosts = append(hosts, transferToHost(resp.DataList)...)

		if len(resp.DataList) < int(core.DefaultMaxPageLimit) {
			break
		}

		req.PageIndex++
	}

	return hosts, nil
}

func transferToHost(devices []caiche.Device) []define.RecycleHostTable {
	hosts := make([]define.RecycleHostTable, 0, len(devices))
	for _, device := range devices {
		data := define.RecycleHostTable{
			AssetID:      converter.ValToPtr(device.SvrAssetId),
			InnerIP:      converter.ValToPtr(device.ServerLanIP),
			Module:       converter.ValToPtr(device.Module),
			AbolishPhase: converter.ValToPtr(device.AbolishPhase),
		}

		hosts = append(hosts, data)
	}

	return hosts
}

func diff(dbHosts []define.RecycleHostTable, caiCheHosts []define.RecycleHostTable) (
	[]define.RecycleHostTable, []define.RecycleHostTable, []string) {

	dbMap := make(map[string]define.RecycleHostTable, len(dbHosts))
	for _, one := range dbHosts {
		dbMap[converter.PtrToVal(one.AssetID)] = one
	}

	create := make([]define.RecycleHostTable, 0)
	update := make([]define.RecycleHostTable, 0)
	for _, newHost := range caiCheHosts {
		dbHost, exist := dbMap[converter.PtrToVal(newHost.AssetID)]
		if !exist {
			create = append(create, newHost)
			continue
		}

		delete(dbMap, converter.PtrToVal(newHost.AssetID))

		if isChange(newHost, dbHost) {
			newHost.ID = dbHost.ID
			update = append(update, newHost)
		}
	}

	deleteIDs := make([]string, 0)
	for _, one := range dbMap {
		deleteIDs = append(deleteIDs, one.ID)
	}

	return create, update, deleteIDs
}

func isChange(new, old define.RecycleHostTable) bool {
	if *new.InnerIP != *old.InnerIP {
		return true
	}

	if *new.Module != *old.Module {
		return true
	}

	if *new.AbolishPhase != *old.AbolishPhase {
		return true
	}

	return false
}
