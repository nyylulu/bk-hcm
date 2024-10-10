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

package clean

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	daotypes "hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// CleanOption clean option
type CleanOption struct {
	Vendor        enumor.Vendor `json:"vendor" validate:"required"`
	BillYear      int           `json:"bill_year" validate:"required"`
	BillMonth     int           `json:"bill_month" validate:"required"`
	MainAccountID string        `json:"main_account_id" validate:"required"`
}

var _ action.Action = new(CleanAction)
var _ action.ParameterAction = new(CleanAction)

// CleanAction define sync action
type CleanAction struct{}

// ParameterNew return request params.
func (act CleanAction) ParameterNew() interface{} {
	return new(CleanOption)
}

// Name return action name
func (act CleanAction) Name() enumor.ActionName {
	return enumor.ActionObsClean
}

// Run run sync
func (act CleanAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*CleanOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	expressions := []*filter.AtomRule{
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
		tools.RuleEqual("main_account_id", opt.MainAccountID),
	}
	filter := tools.ExpressionAnd(expressions...)

	switch opt.Vendor {
	case enumor.HuaWei:
		if err := act.doHuaweiClean(kt.Kit(), filter); err != nil {
			return nil, err
		}
		return nil, nil
	case enumor.Gcp:
		if err := act.doGcpClean(kt.Kit(), filter); err != nil {
			return nil, err
		}
		return nil, nil
	case enumor.Aws:
		if err := act.doAwsClean(kt.Kit(), filter); err != nil {
			return nil, err
		}
		return nil, nil
	case enumor.Zenlayer:
		if err := act.doZenlayerClean(kt.Kit(), filter); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported vendor %s", opt.Vendor)
	}
}

func (act CleanAction) doHuaweiClean(kt *kit.Kit, filter *filter.Expression) error {
	for {
		result, err := actcli.GetObsDaoSet().OBSBillItemHuawei().List(kt, &daotypes.ListOption{
			Filter: filter,
			Page: &core.BasePage{
				Count: true,
			},
		})
		if err != nil {
			logs.Warnf("count huawei obs bill item failed, err %s, rid: %s", err.Error(), kt.Rid)
			return fmt.Errorf("count huawei obs bill item failed, err %s", err.Error())
		}
		if result.Count == nil {
			logs.Warnf("count huawei obs bill item failed, empty count, resp %v rid: %s", result, kt.Rid)
			return fmt.Errorf("count huawei obs bill item failed, empty count, resp %v", result)
		}
		logs.Infof("found huawei obs bill item count %d, rid: %s", *result.Count, kt.Rid)
		if *result.Count > 0 {
			_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt,
				func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
					if err := actcli.GetObsDaoSet().OBSBillItemHuawei().DeleteWithTx(
						kt, txn, filter, uint64(core.DefaultMaxPageLimit)); err != nil {
						logs.Warnf("delete huawei obs bill item by filter %s failed, err %s, rid: %s", filter, kt.Rid)
					}
					return nil, nil
				})
			if err != nil {
				return err
			}
			logs.Infof("successfully clean huawei obs bill item count %d, rid: %s", core.DefaultMaxPageLimit, kt.Rid)
			continue
		}
		break
	}
	return nil
}

func (act CleanAction) doGcpClean(kt *kit.Kit, filter *filter.Expression) error {
	for {
		result, err := actcli.GetObsDaoSet().OBSBillItemGcp().List(kt, &daotypes.ListOption{
			Filter: filter,
			Page: &core.BasePage{
				Count: true,
			},
		})
		if err != nil {
			logs.Warnf("count gcp obs bill item failed, err %s, rid: %s", err.Error(), kt.Rid)
			return fmt.Errorf("count gcp obs bill item failed, err %s", err.Error())
		}

		logs.Infof("found gcp obs bill item count %d, rid: %s", result.Count, kt.Rid)
		if result.Count > 0 {
			_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt,
				func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
					if err := actcli.GetObsDaoSet().OBSBillItemGcp().DeleteWithTx(
						kt, txn, filter, uint64(core.DefaultMaxPageLimit)); err != nil {
						logs.Warnf("delete gcp obs bill item by filter %s failed, err %s, rid: %s", filter, kt.Rid)
					}
					return nil, nil
				})
			if err != nil {
				return err
			}
			logs.Infof("successfully clean gcp obs bill item count %d, rid: %s", core.DefaultMaxPageLimit, kt.Rid)
			continue
		}
		break
	}
	return nil
}

func (act CleanAction) doAwsClean(kt *kit.Kit, filter *filter.Expression) error {
	for {
		result, err := actcli.GetObsDaoSet().OBSBillItemAws().List(kt, &daotypes.ListOption{
			Filter: filter,
			Page: &core.BasePage{
				Count: true,
			},
		})
		if err != nil {
			logs.Warnf("count aws obs bill item failed, err %s, rid: %s", err.Error(), kt.Rid)
			return fmt.Errorf("count aws obs bill item failed, err %s", err.Error())
		}

		logs.Infof("found aws obs bill item count %d, rid: %s", result.Count, kt.Rid)
		if result.Count > 0 {
			_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt,
				func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
					if err := actcli.GetObsDaoSet().OBSBillItemAws().DeleteWithTx(
						kt, txn, filter, uint64(core.DefaultMaxPageLimit)); err != nil {
						logs.Warnf("delete aws obs bill item by filter %s failed, err %s, rid: %s", filter, kt.Rid)
					}
					return nil, nil
				})
			if err != nil {
				return err
			}
			logs.Infof("successfully clean aws obs bill item count %d, rid: %s", core.DefaultMaxPageLimit, kt.Rid)
			continue
		}
		break
	}
	return nil
}

func (act CleanAction) doZenlayerClean(kt *kit.Kit, filter *filter.Expression) error {
	for {
		result, err := actcli.GetObsDaoSet().OBSBillItemZenlayer().List(kt, &daotypes.ListOption{
			Filter: filter,
			Page: &core.BasePage{
				Count: true,
			},
		})
		if err != nil {
			logs.Warnf("count zenlayer obs bill item failed, err %s, rid: %s", err.Error(), kt.Rid)
			return fmt.Errorf("count zenlayer obs bill item failed, err %s", err.Error())
		}

		logs.Infof("found zenlayer obs bill item count %d, rid: %s", result.Count, kt.Rid)
		if result.Count > 0 {
			_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt,
				func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
					if err := actcli.GetObsDaoSet().OBSBillItemZenlayer().DeleteWithTx(
						kt, txn, filter, uint64(core.DefaultMaxPageLimit)); err != nil {
						logs.Warnf("delete zenlayer obs bill item by filter %s failed, err %s, rid: %s", filter, kt.Rid)
					}
					return nil, nil
				})
			if err != nil {
				return err
			}
			logs.Infof("successfully clean zenlayer obs bill item count %d", core.DefaultMaxPageLimit)
			continue
		}
		break
	}
	return nil
}
