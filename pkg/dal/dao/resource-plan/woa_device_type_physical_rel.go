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
	resourceplan "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	wd "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// WoaDeviceTypePhysicalRelInterface only used for woa device type physical rel interface.
type WoaDeviceTypePhysicalRelInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []wd.WoaDeviceTypePhysicalRelTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, model *wd.WoaDeviceTypePhysicalRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*resourceplan.WoaDeviceTypePhysicalRelListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ WoaDeviceTypePhysicalRelInterface = new(WoaDeviceTypePhysicalRelDao)

// WoaDeviceTypePhysicalRelDao woa device type physical rel dao.
type WoaDeviceTypePhysicalRelDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create woa device type physical rel with transaction.
func (dao WoaDeviceTypePhysicalRelDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []wd.WoaDeviceTypePhysicalRelTable) ([]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := dao.IDGen.Batch(kt, table.WoaDeviceTypePhysicalRelTable, len(models))
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
		wd.WoaDeviceTypePhysicalRelColumns.ColumnExpr(), wd.WoaDeviceTypePhysicalRelColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("create woa device type physical rel failed, err: %v, models: %+v, rid: %v", err, models, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// UpdateWithTx update woa device type physical rel with transaction.
func (dao WoaDeviceTypePhysicalRelDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
	model *wd.WoaDeviceTypePhysicalRelTable) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}
	if err := model.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}
	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}
	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	effected, err := dao.Orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.Errorf("update woa device type physical rel failed, filter: %v, err: %v, rid: %v",
			expr, err, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.Errorf("update woa device type physical rel, but record not found, filter: %v, rid: %v",
			expr, kt.Rid)
		return errf.New(errf.RecordNotFound, "record not found")
	}

	return nil
}

// List list woa device type physical rel.
func (dao WoaDeviceTypePhysicalRelDao) List(kt *kit.Kit, opt *types.ListOption) (
	*resourceplan.WoaDeviceTypePhysicalRelListResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	exprOpt := filter.NewExprOption(filter.RuleFields(wd.WoaDeviceTypePhysicalRelColumns.ColumnTypes()))
	if err := opt.Validate(exprOpt, core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.WoaDeviceTypePhysicalRelTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count woa device type physical rel failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &resourceplan.WoaDeviceTypePhysicalRelListResult{Count: count}, nil
	}
	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, wd.WoaDeviceTypePhysicalRelColumns.FieldsNamedExpr(opt.Fields),
		table.WoaDeviceTypePhysicalRelTable, whereExpr, pageExpr)

	details := make([]wd.WoaDeviceTypePhysicalRelTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &resourceplan.WoaDeviceTypePhysicalRelListResult{Details: details}, nil
}

// DeleteWithTx delete woa device type physical rel with transaction.
func (dao WoaDeviceTypePhysicalRelDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.WoaDeviceTypePhysicalRelTable, whereExpr)

	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete woa device type physical rel failed, err: %v, filter: %v, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
