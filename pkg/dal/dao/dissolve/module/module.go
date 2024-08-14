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
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/dissolve/module"
	"hcm/pkg/dal/table"
	define "hcm/pkg/dal/table/dissolve/module"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// RecycleModule defines recycle module dao operations.
type RecycleModule interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []define.RecycleModuleTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, module *define.RecycleModuleTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*module.ListRecycleModuleDetails,
		error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ RecycleModule = new(Dao)

// Dao recycle module dao.
type Dao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewRecycleModuleDao create a recycle Module dao.
func NewRecycleModuleDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) RecycleModule {
	return &Dao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// CreateWithTx create recycle module with transaction.
func (d *Dao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, modules []define.RecycleModuleTable) ([]string, error) {
	if len(modules) == 0 {
		return nil, errf.New(errf.InvalidParameter, "modules to create cannot be empty")
	}

	ids, err := d.idGen.Batch(kt, table.RecycleModuleInfo, len(modules))
	if err != nil {
		return nil, err
	}

	for idx := range modules {
		modules[idx].ID = ids[idx]
		modules[idx].Creator = kt.User
		if err = modules[idx].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.RecycleModuleInfo,
		define.RecycleModuleColumns.ColumnExpr(), define.RecycleModuleColumns.ColonNameExpr())

	err = d.orm.Txn(tx).BulkInsert(kt.Ctx, sql, modules)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", table.RecycleModuleInfo, err)
	}

	return ids, nil
}

// UpdateWithTx update recycle module with transaction.
func (d *Dao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, module *define.RecycleModuleTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	module.Reviser = kt.User
	if err := module.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(module, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, module.TableName(), setExpr, whereExpr)

	effected, err := d.orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.ErrorJson("update resource recycle module failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update resource recycle module, but data not found, filter: %v, rid: %v", expr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List recycle Modules.
func (d *Dao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
	*module.ListRecycleModuleDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list recycle module options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(define.RecycleModuleColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereOpt := tools.DefaultSqlWhereOption
	if len(whereOpts) != 0 && whereOpts[0] != nil {
		err := whereOpts[0].Validate()
		if err != nil {
			return nil, err
		}
		whereOpt = whereOpts[0]
	}

	if opt.Filter == nil {
		opt.Filter = tools.AllExpression()
	}
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.RecycleModuleInfo, whereExpr)

		count, err := d.orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count recycle modules failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &module.ListRecycleModuleDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, define.RecycleModuleColumns.FieldsNamedExpr(opt.Fields),
		table.RecycleModuleInfo, whereExpr, pageExpr)

	details := make([]define.RecycleModuleTable, 0)
	if err = d.orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &module.ListRecycleModuleDetails{Details: details}, nil
}

// DeleteWithTx delete recycle module with transaction.
func (d *Dao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.RecycleModuleInfo, whereExpr)
	if _, err = d.orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete recycle module failed, err: %v, filter: %v, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
