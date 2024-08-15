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

package module

import (
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/dissolve/module"
	define "hcm/pkg/dal/table/dissolve/module"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
)

// RecycledModule provides interface for operations of recycle module.
type RecycledModule interface {
	Create(kt *kit.Kit, modules []define.RecycleModuleTable) ([]string, error)
	Update(kt *kit.Kit, module *define.RecycleModuleTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*module.ListRecycleModuleDetails, error)
	Delete(kt *kit.Kit, ids []string) error
}

type logics struct {
	dao dao.Set
}

// New create recycle module logics.
func New(dao dao.Set) RecycledModule {
	return &logics{dao: dao}
}

// Create recycle module.
func (l *logics) Create(kt *kit.Kit, modules []define.RecycleModuleTable) ([]string, error) {
	result, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		res, err := l.dao.RecycleModule().CreateWithTx(kt, txn, modules)
		if err != nil {
			logs.Errorf("create recycle module failed, err: %v, data: %+v, rid: %s", err, modules, kt.Rid)
			return nil, err
		}

		return res, nil
	})

	if err != nil {
		return nil, err
	}

	return result.([]string), nil
}

// Update recycle module.
func (l *logics) Update(kt *kit.Kit, module *define.RecycleModuleTable) error {
	_, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		err := l.dao.RecycleModule().UpdateWithTx(kt, txn, tools.EqualExpression("id", module.ID), module)
		if err != nil {
			logs.Errorf("update recycle module failed, err: %v, data: %+v, rid: %s", err, module, kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}

// List recycle module.
func (l *logics) List(kt *kit.Kit, opt *types.ListOption) (
	*module.ListRecycleModuleDetails, error) {

	return l.dao.RecycleModule().List(kt, opt)
}

// Delete recycle module.
func (l *logics) Delete(kt *kit.Kit, ids []string) error {
	_, err := l.dao.Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		err := l.dao.RecycleModule().DeleteWithTx(kt, txn, tools.ContainersExpression("id", ids))
		if err != nil {
			logs.Errorf("delete recycle module failed, err: %v, ids: %+v, rid: %s", err, ids, kt.Rid)
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}
