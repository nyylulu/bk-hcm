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

// Package obs ...
package obs

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typesobs "hcm/pkg/dal/dao/types/obs"
	"hcm/pkg/dal/table"
	tableobs "hcm/pkg/dal/table/obs"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// OBSBillItemGcp only used for interface.
type OBSBillItemGcp interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, items []*tableobs.OBSBillItemGcp) ([]string, error)
	List(kt *kit.Kit, opt *types.ListOption) (*typesobs.ListOBSBillItemGcpDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, f *filter.Expression, limit uint64) error
}

// OBSBillItemGcpDao account bill item dao
type OBSBillItemGcpDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx create account bill item with tx.
func (o OBSBillItemGcpDao) CreateWithTx(
	kt *kit.Kit, tx *sqlx.Tx, models []*tableobs.OBSBillItemGcp) (
	[]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := o.IDGen.Batch(kt, models[0].TableName(), len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		models[index].ID = ids[index]

		if err = model.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		tableobs.OBSBillItemGcpColumns.ColumnExpr(), tableobs.OBSBillItemGcpColumns.ColonNameExpr())

	if err = o.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// List get account bill item list.
func (o OBSBillItemGcpDao) List(kt *kit.Kit, opt *types.ListOption) (
	*typesobs.ListOBSBillItemGcpDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account bill item options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableobs.OBSBillItemGcpColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.OBSBillGcpItemTable, whereExpr)
		count, err := o.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count account bill item failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesobs.ListOBSBillItemGcpDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableobs.OBSBillItemGcpColumns.FieldsNamedExpr(opt.Fields),
		table.OBSBillGcpItemTable, whereExpr, pageExpr)

	details := make([]tableobs.OBSBillItemGcp, 0)
	if err = o.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}
	return &typesobs.ListOBSBillItemGcpDetails{Details: details}, nil
}

// DeleteWithTx delete account bill item with tx.
func (a OBSBillItemGcpDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, f *filter.Expression, limit uint64) error {
	if f == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := f.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	idSql := fmt.Sprintf(`SELECT id FROM %s %s LIMIT %d`, table.OBSBillGcpItemTable, whereExpr, limit)
	preDetails := make([]tableobs.OBSBillItemGcp, 0)
	if err = a.Orm.Do().Select(kt.Ctx, &preDetails, idSql, whereValue); err != nil {
		return err
	}
	detailIDs := make([]string, 0, len(preDetails))
	for _, detail := range preDetails {
		detailIDs = append(detailIDs, detail.ID)
	}

	if len(detailIDs) == 0 {
		return nil
	}

	sql := fmt.Sprintf(`DELETE FROM %s WHERE id IN (:ids) LIMIT %d`, table.OBSBillGcpItemTable, limit)

	if _, err = a.Orm.Txn(tx).Delete(kt.Ctx, sql, map[string]interface{}{"ids": detailIDs}); err != nil {
		logs.ErrorJson("delete obs gcp bill item failed, err: %v, filter: %s, limit: %d, rid: %s",
			err, f, limit, kt.Rid)
		return err
	}
	return nil
}
