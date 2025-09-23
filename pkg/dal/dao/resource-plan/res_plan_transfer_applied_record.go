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

package resplan

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgen "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	resplan "hcm/pkg/dal/dao/types/resource-plan"
	"hcm/pkg/dal/table"
	rptar "hcm/pkg/dal/table/resource-plan/res-plan-transfer-applied-record"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// TransferAppliedRecordInterface 转移额度执行记录表操作接口
type TransferAppliedRecordInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []rptar.ResPlanTransferAppliedRecordTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
		model *rptar.ResPlanTransferAppliedRecordTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*resplan.ResPlanTransferAppliedRecordListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	SumUsedTransferAppliedRecord(kt *kit.Kit, opt *types.ListOption) (*resplan.SumTransferAppliedRecord, error)
}

var _ TransferAppliedRecordInterface = new(TransferAppliedRecordDao)

// TransferAppliedRecordDao 转移额度执行记录DAO实现
type TransferAppliedRecordDao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx 创建转移额度执行记录(带事务)
func (d TransferAppliedRecordDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []rptar.ResPlanTransferAppliedRecordTable) ([]string, error) {

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

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`, models[0].TableName(),
		rptar.ResPlanTransferAppliedRecordColumns.ColumnExpr(),
		rptar.ResPlanTransferAppliedRecordColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// UpdateWithTx 更新转移额度执行记录(带事务)
func (d TransferAppliedRecordDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression,
	model *rptar.ResPlanTransferAppliedRecordTable) error {
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

	effected, err := d.Orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.Errorf("update transfer applied record failed, filter: %v, err: %v, rid: %s", filterExpr, err, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.Errorf("update transfer applied record, but record not found, filter: %v, rid: %s", filterExpr, kt.Rid)
	}

	return nil
}

// List 查询转移额度执行记录列表
func (d TransferAppliedRecordDao) List(kt *kit.Kit, opt *types.ListOption) (
	*resplan.ResPlanTransferAppliedRecordListResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(
		rptar.ResPlanTransferAppliedRecordColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ResPlanTransferAppliedRecordTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count transfer applied record failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &resplan.ResPlanTransferAppliedRecordListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, rptar.ResPlanTransferAppliedRecordColumns.FieldsNamedExpr(opt.Fields),
		table.ResPlanTransferAppliedRecordTable, whereExpr, pageExpr)

	details := make([]rptar.ResPlanTransferAppliedRecordTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &resplan.ResPlanTransferAppliedRecordListResult{Count: 0, Details: details}, nil
}

// DeleteWithTx 删除转移额度执行记录(带事务)
func (d TransferAppliedRecordDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ResPlanTransferAppliedRecordTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete transfer applied record failed, err: %v, filter: %v, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// SumUsedTransferAppliedRecord 查询已使用的额度
func (d TransferAppliedRecordDao) SumUsedTransferAppliedRecord(kt *kit.Kit, opt *types.ListOption) (
	*resplan.SumTransferAppliedRecord, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(
		rptar.ResPlanTransferAppliedRecordColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT IFNULL(SUM(applied_core), 0) AS sum_applied_core, 
       IFNULL(SUM(expected_core), 0) AS sum_expected_core FROM %s %s`,
		table.ResPlanTransferAppliedRecordTable, whereExpr)

	details := make([]*resplan.SumTransferAppliedRecord, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return details[0], nil
}
