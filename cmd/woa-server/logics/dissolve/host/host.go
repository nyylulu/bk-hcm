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
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/dissolve/host"
	define "hcm/pkg/dal/table/dissolve/host"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
)

// RecycledHost provides interface for operations of recycle host.
type RecycledHost interface {
	Create(kt *kit.Kit, hosts []define.RecycleHostTable) ([]string, error)
	Update(kt *kit.Kit, host *define.RecycleHostTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*host.ListRecycleHostDetails,
		error)
	Delete(kt *kit.Kit, ids []string) error
}

type logics struct {
	dao dao.Set
}

// New create recycle host logics.
func New(dao dao.Set) RecycledHost {
	return &logics{dao: dao}
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
