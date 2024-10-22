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

// Package rollingserver ...
package rollingserver

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rtypes "hcm/pkg/dal/dao/types/rolling-server"
	"hcm/pkg/dal/table"
	tablers "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// RollingQuotaOffsetInterface only used for rolling quota offset interface.
type RollingQuotaOffsetInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tablers.RollingQuotaOffsetTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tablers.RollingQuotaOffsetTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*rtypes.RollingQuotaOffsetListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ RollingQuotaOffsetInterface = new(RollingQuotaOffsetDao)

// RollingQuotaOffsetDao rolling quota offset RollingQuotaOffsetDao.
type RollingQuotaOffsetDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx create rolling quota offset with tx.
func (d RollingQuotaOffsetDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tablers.RollingQuotaOffsetTable) (
	[]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := d.IDGen.Batch(kt, models[0].TableName(), len(models))
	if err != nil {
		return nil, err
	}

	for index := range models {
		models[index].ID = ids[index]

		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		tablers.RollingQuotaOffsetColumns.ColumnExpr(), tablers.RollingQuotaOffsetColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// Update update rolling quota offset.
func (d RollingQuotaOffsetDao) Update(kt *kit.Kit, filterExpr *filter.Expression,
	model *tablers.RollingQuotaOffsetTable) error {

	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = d.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := d.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update rolling quota offset failed, filter: %v, err: %v, rid: %v",
				filterExpr, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update rolling quota offset, but record not found, filter: %v, rid: %v",
				filterExpr, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List get rolling quota offset list.
func (d RollingQuotaOffsetDao) List(kt *kit.Kit, opt *types.ListOption) (*rtypes.RollingQuotaOffsetListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list rolling quota offset options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablers.RollingQuotaOffsetColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.RollingQuotaOffsetTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count rolling quota offset failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &rtypes.RollingQuotaOffsetListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablers.RollingQuotaOffsetColumns.FieldsNamedExpr(opt.Fields),
		table.RollingQuotaOffsetTable, whereExpr, pageExpr)

	details := make([]tablers.RollingQuotaOffsetTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &rtypes.RollingQuotaOffsetListResult{Count: 0, Details: details}, nil
}

// DeleteWithTx delete rolling quota offset with tx.
func (d RollingQuotaOffsetDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.RollingQuotaOffsetTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete rolling quota offset failed, err: %v, filter: %v, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
