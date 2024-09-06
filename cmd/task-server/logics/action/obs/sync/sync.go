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

package sync

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/shopspring/decimal"
)

var defaultRetryTimes = 10

// SyncOption option for sync action
type SyncOption struct {
	Vendor        enumor.Vendor `json:"vendor" validate:"required"`
	BillYear      int           `json:"bill_year" validate:"required"`
	BillMonth     int           `json:"bill_month" validate:"required"`
	MainAccountID string        `json:"main_account_id" validate:"required"`
	Start         uint64        `json:"start" validate:"required"`
	Limit         uint64        `json:"limit" validate:"required"`
}

var _ action.Action = new(SyncAction)
var _ action.ParameterAction = new(SyncAction)

// SyncAction define sync action
type SyncAction struct{}

// ParameterNew return request params.
func (act SyncAction) ParameterNew() interface{} {
	return new(SyncOption)
}

// Name return action name
func (act SyncAction) Name() enumor.ActionName {
	return enumor.ActionObsSync
}

// Run run sync
func (act SyncAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*SyncOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	switch opt.Vendor {
	case enumor.HuaWei:
		mainAccount, err := act.getHuaweiMainAccount(kt.Kit(), opt.MainAccountID)
		if err != nil {
			return nil, fmt.Errorf("get main account failed, err %s", err.Error())
		}
		if err := act.doBatchSyncHuaweiBillitem(kt.Kit(), mainAccount, opt); err != nil {
			return nil, err
		}
		return nil, nil
	case enumor.Aws:
		mainAccount, err := act.getAwsMainAccount(kt.Kit(), opt.MainAccountID)
		if err != nil {
			return nil, fmt.Errorf("get main account failed, err %s", err.Error())
		}
		if err := act.doBatchSyncAwsBillitem(kt.Kit(), mainAccount, opt); err != nil {
			return nil, err
		}
		return nil, nil
	case enumor.Gcp:
		mainAccount, err := act.getGcpMainAccount(kt.Kit(), opt.MainAccountID)
		if err != nil {
			return nil, fmt.Errorf("get main account failed, err %s", err.Error())
		}
		if err := act.doBatchSyncGcpBillitem(kt.Kit(), mainAccount, opt); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported obs vendor %s", opt.Vendor)
	}
}

func (act *SyncAction) getExchangeRate(
	kt *kit.Kit, fromCurrency, toCurrency enumor.CurrencyCode, billYear, billMonth int) (*decimal.Decimal, error) {
	expressions := []*filter.AtomRule{
		tools.RuleEqual("from_currency", fromCurrency),
		tools.RuleEqual("to_currency", toCurrency),
		tools.RuleEqual("year", billYear),
		tools.RuleEqual("month", billMonth),
	}
	result, err := actcli.GetDataService().Global.Bill.ListExchangeRate(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get exchange rate from %s to %s in %d-%d failed, err %s",
			fromCurrency, toCurrency, billYear, billMonth, err.Error())
	}
	if len(result.Details) == 0 {
		logs.Infof("get no exchange rate from %s to %s in %d-%d, rid %s",
			fromCurrency, toCurrency, billYear, billMonth, kt.Rid)
		return nil, nil
	}
	if len(result.Details) != 1 {
		logs.Infof("get invalid resp length from exchange rate from %s to %s in %d-%d, resp %v, rid %s",
			fromCurrency, toCurrency, billYear, billMonth, result.Details, kt.Rid)
		return nil, fmt.Errorf("get invalid resp length from exchange rate from %s to %s in %d-%d, resp %v",
			fromCurrency, toCurrency, billYear, billMonth, result.Details)
	}
	return result.Details[0].ExchangeRate, nil
}
