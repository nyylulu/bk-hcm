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
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	rstable "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// RollingReturnedRecordInterface only used for rolling returned record interface.
type RollingReturnedRecordInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []rstable.RollingReturnedRecord) ([]string, error)
	Update(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, model *rstable.RollingReturnedRecord) error
	List(kt *kit.Kit, opt *types.ListOption) (*rsproto.RollingReturnedRecordListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	GetReturnedSumDeliveredCore(kt *kit.Kit, opt *types.ListOption) (*rsproto.RollingCpuCoreSummaryItem, error)
}

var _ RollingReturnedRecordInterface = new(RollingReturnedRecordDao)

// RollingReturnedRecordDao dao.
type RollingReturnedRecordDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create rolling returned record with tx.
func (d RollingReturnedRecordDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []rstable.RollingReturnedRecord) (
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
		rstable.RollingReturnedRecordColumns.ColumnExpr(), rstable.RollingReturnedRecordColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert table %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert table %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// Update update rolling returned record.
func (d RollingReturnedRecordDao) Update(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression,
	model *rstable.RollingReturnedRecord) error {

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

	ignoredFields := append(types.DefaultIgnoredFields, "updated_at")
	opts := utils.NewFieldOptions().AddIgnoredFields(ignoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)
	updateValue := tools.MapMerge(toUpdate, whereValue)
	effected, err := d.Orm.Txn(tx).Update(kt.Ctx, sql, updateValue)
	if err != nil {
		logs.ErrorJson("update rolling returned record failed, sql: %s, err: %v, updateValue: %+v, rid: %s",
			sql, err, updateValue, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update rolling returned record, but record not found, sql: %s, updateValue: %+v, rid: %s",
			sql, updateValue, kt.Rid)
	}
	return nil
}

// List get rolling returned record list.
func (d RollingReturnedRecordDao) List(kt *kit.Kit, opt *types.ListOption) (
	*rsproto.RollingReturnedRecordListResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list rolling returned record options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(rstable.RollingReturnedRecordColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.RollingReturnedRecordTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count rolling returned recore failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &rsproto.RollingReturnedRecordListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, rstable.RollingReturnedRecordColumns.FieldsNamedExpr(opt.Fields),
		table.RollingReturnedRecordTable, whereExpr, pageExpr)

	details := make([]*rstable.RollingReturnedRecord, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &rsproto.RollingReturnedRecordListResult{Count: 0, Details: details}, nil
}

// DeleteWithTx delete rolling returned record with tx.
func (d RollingReturnedRecordDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.RollingReturnedRecordTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete rolling returned record failed, err: %v, whereValue: %+v, rid: %s",
			err, whereValue, kt.Rid)
		return err
	}

	return nil
}

// GetReturnedSumDeliveredCore get returned sum match applied core.
func (d RollingReturnedRecordDao) GetReturnedSumDeliveredCore(kt *kit.Kit, opt *types.ListOption) (
	*rsproto.RollingCpuCoreSummaryItem, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "get rolling returned sum match  core options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(rstable.RollingReturnedRecordColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	result := make([]*rsproto.RollingCpuCoreSummaryItem, 0)
	sql := fmt.Sprintf(`SELECT SUM(match_applied_core) AS sum_returned_applied_core FROM %s %s`,
		table.RollingReturnedRecordTable, whereExpr)
	if err = d.Orm.Do().Select(kt.Ctx, &result, sql, whereValue); err != nil {
		logs.ErrorJson("get rolling returned sum match applied core failed, err: %v, sql: %s, whereValue: %+v, rid: %s",
			err, sql, whereValue, kt.Rid)
		return nil, err
	}
	// 空数据
	if len(result) == 0 {
		return &rsproto.RollingCpuCoreSummaryItem{}, nil
	}
	return result[0], nil
}
