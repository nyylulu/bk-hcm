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

// Package resplan ...
package resplan

import (
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	wdt "hcm/pkg/dal/table/resource-plan/woa-device-type"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// WoaDeviceTypeInterface only used for woa device type interface.
type WoaDeviceTypeInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []wdt.WoaDeviceTypeTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *wdt.WoaDeviceTypeTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*rpproto.WoaDeviceTypeListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	// GetDeviceTypeMap get device type table mapping.
	GetDeviceTypeMap(kt *kit.Kit, expr *filter.Expression) (map[string]wdt.WoaDeviceTypeTable, error)
	// GetDeviceClassList get device class list.
	GetDeviceClassList(kt *kit.Kit, expr *filter.Expression) ([]string, error)
}

var _ WoaDeviceTypeInterface = new(WoaDeviceTypeDao)

// WoaDeviceTypeDao woa device type WoaDeviceTypeDao.
type WoaDeviceTypeDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create woa device type with tx.
func (d WoaDeviceTypeDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []wdt.WoaDeviceTypeTable) ([]string, error) {
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
		wdt.WoaDeviceTypeColumns.ColumnExpr(), wdt.WoaDeviceTypeColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// Update update woa device type.
func (d WoaDeviceTypeDao) Update(kt *kit.Kit, filterExpr *filter.Expression, model *wdt.WoaDeviceTypeTable) error {
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
			logs.ErrorJson("update woa device type failed, filter: %v, err: %v, rid: %v", filterExpr, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update woa device type, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List get woa device type list.
func (d WoaDeviceTypeDao) List(kt *kit.Kit, opt *types.ListOption) (*rpproto.WoaDeviceTypeListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list woa device type options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(wdt.WoaDeviceTypeColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.WoaDeviceTypeTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count woa device type failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &rpproto.WoaDeviceTypeListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, wdt.WoaDeviceTypeColumns.FieldsNamedExpr(opt.Fields),
		table.WoaDeviceTypeTable, whereExpr, pageExpr)

	details := make([]wdt.WoaDeviceTypeTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &rpproto.WoaDeviceTypeListResult{Count: 0, Details: details}, nil
}

// DeleteWithTx delete woa device type with tx.
func (d WoaDeviceTypeDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.WoaDeviceTypeTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete woa device type failed, err: %v, filter: %v, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// GetDeviceTypeMap get device type table mapping.
func (d WoaDeviceTypeDao) GetDeviceTypeMap(kt *kit.Kit, expr *filter.Expression) (
	map[string]wdt.WoaDeviceTypeTable, error) {

	if expr == nil {
		return nil, errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s`,
		wdt.WoaDeviceTypeColumns.FieldsNamedExpr(nil), table.WoaDeviceTypeTable, whereExpr)
	details := make([]wdt.WoaDeviceTypeTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	deviceTypeMap := make(map[string]wdt.WoaDeviceTypeTable)
	for _, detail := range details {
		deviceTypeMap[detail.DeviceType] = detail
	}

	return deviceTypeMap, nil
}

// GetDeviceClassList get device class list.
func (d WoaDeviceTypeDao) GetDeviceClassList(kt *kit.Kit, expr *filter.Expression) ([]string, error) {
	if expr == nil {
		return nil, errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT DISTINCT device_class FROM %s %s`, table.WoaDeviceTypeTable, whereExpr)
	details := make([]wdt.WoaDeviceTypeTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	result := make([]string, 0, len(details))
	for _, detail := range details {
		result = append(result, detail.DeviceClass)
	}

	return result, nil
}
