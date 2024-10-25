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
	rtypes "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableobs "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// OBSBillItemRolling only used for interface.
type OBSBillItemRolling interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, items []tableobs.OBSBillItemRolling) ([]string, error)
	List(kt *kit.Kit, opt *types.ListOption) (*rtypes.RollingBillListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, f *filter.Expression) error
}

// OBSBillItemRollingDao rolling bill item dao
type OBSBillItemRollingDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx create rolling bill item with tx.
func (o OBSBillItemRollingDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableobs.OBSBillItemRolling) ([]string,
	error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := o.IDGen.Batch(kt, models[0].TableName(), len(models))
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
		tableobs.OBSBillItemRollingColumns.ColumnExpr(), tableobs.OBSBillItemRollingColumns.ColonNameExpr())

	if err = o.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// List get rolling bill item list.
func (o OBSBillItemRollingDao) List(kt *kit.Kit, opt *types.ListOption) (*rtypes.RollingBillListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list rolling bill item options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableobs.OBSBillItemRollingColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.OBSBillRollingItemTable, whereExpr)
		count, err := o.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count rolling bill item failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &rtypes.RollingBillListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableobs.OBSBillItemRollingColumns.FieldsNamedExpr(opt.Fields),
		table.OBSBillRollingItemTable, whereExpr, pageExpr)

	details := make([]*tableobs.OBSBillItemRolling, 0)
	if err = o.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}
	return &rtypes.RollingBillListResult{Details: details}, nil
}

// DeleteWithTx delete rolling bill item with tx.
func (a OBSBillItemRollingDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.OBSBillRollingItemTable, whereExpr)

	if _, err = a.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete rolling bill failed, err: %v, whereValue: %+v, rid: %s", err, whereValue, kt.Rid)
		return err
	}

	return nil
}
