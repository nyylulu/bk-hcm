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
	"math/rand"
	"time"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	accountsetcore "hcm/pkg/api/core/account-set"
	accountsetproto "hcm/pkg/api/data-service/account-set"
	databill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tableobs "hcm/pkg/dal/table/obs"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
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

func (act SyncAction) doBatchSyncHuaweiBillitem(kt *kit.Kit,
	mainAccount *accountsetproto.MainAccountGetResult[accountsetcore.HuaWeiMainAccountExtension],
	syncOpt *SyncOption) error {
	for start := syncOpt.Start; start < syncOpt.Start+syncOpt.Limit; start = start + uint64(core.DefaultMaxPageLimit) {
		var err error
		for retry := 0; retry < defaultRetryTimes; retry++ {
			if err = act.doSyncHuaweiBillItem(
				kt, mainAccount, syncOpt, start, uint64(core.DefaultMaxPageLimit)); err != nil {

				logs.Warnf("do sync huawei bill %v, start %d, limit %d, retry %d, rid %s",
					syncOpt, start, uint64(core.DefaultMaxPageLimit), retry+1, kt.Rid)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))
				continue
			}
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (act SyncAction) doSyncHuaweiBillItem(
	kt *kit.Kit,
	mainAccount *accountsetproto.MainAccountGetResult[accountsetcore.HuaWeiMainAccountExtension],
	syncOpt *SyncOption, start, limit uint64) error {

	flt := tools.ExpressionAnd(
		tools.RuleEqual("vendor", syncOpt.Vendor),
		tools.RuleEqual("bill_year", syncOpt.BillYear),
		tools.RuleEqual("bill_month", syncOpt.BillMonth),
		tools.RuleEqual("main_account_id", syncOpt.MainAccountID),
	)
	comOpt := &databill.ItemCommonOpt{
		Vendor: syncOpt.Vendor,
		Year:   syncOpt.BillYear,
		Month:  syncOpt.BillMonth,
	}
	listReq := &databill.BillItemListReq{
		ItemCommonOpt: comOpt,
		ListReq:       &core.ListReq{Filter: flt, Page: &core.BasePage{Start: uint32(start), Limit: uint(limit)}},
	}
	// 获取分账后的bill item
	result, err := actcli.GetDataService().HuaWei.Bill.ListBillItem(kt, listReq)
	if err != nil {
		logs.Warnf("list huawei bill item by option %v failed, err %s, rid: %s", syncOpt, err.Error(), kt.Rid)
		return err
	}

	if len(result.Details) == 0 {
		logs.Infof("get no bill item for main_account_id %s %d-%d %d-%d",
			syncOpt.MainAccountID, syncOpt.BillYear, syncOpt.BillMonth, start, limit)
		return nil
	}

	// 清理特定的obs数据，此处防止之前有可能插入事务失败导致的脏数据
	setIndex := fmt.Sprintf("%s-%s-%d-%d-%d-%d",
		syncOpt.Vendor, syncOpt.MainAccountID, syncOpt.BillYear, syncOpt.BillMonth, start, limit)
	delExpressions := []*filter.AtomRule{
		tools.RuleEqual("set_index", setIndex),
	}
	deleteFilter := tools.ExpressionAnd(delExpressions...)
	_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := actcli.GetObsDaoSet().OBSBillItemHuawei().DeleteWithTx(
			kt, txn, deleteFilter, uint64(core.DefaultMaxPageLimit)); err != nil {
			logs.Warnf("delete huawei obs bill item by filter %v failed, err %s, rid: %s",
				deleteFilter, err.Error(), kt.Rid)
			return nil, err
		}
		logs.Infof("delete previous obs data for %s successfully", setIndex)
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("delete obs bill txn failed, err %s", err.Error())
	}

	// 进行插入
	finalItems, err := act.convertHuaweiBill(kt, syncOpt, result, setIndex, mainAccount)
	if err != nil {
		logs.Warnf("convert obs huawei bill failed, err %s, rid: %s", err.Error(), kt.Rid)
		return err
	}
	_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if _, err := actcli.GetObsDaoSet().OBSBillItemHuawei().CreateWithTx(kt, txn, finalItems); err != nil {
			logs.Warnf("delete huawei obs bill item by filter %s failed, err %s, rid: %s",
				deleteFilter, err.Error(), kt.Rid)
			return nil, err
		}
		logs.Infof("create obs huawei bill for %s successfully", setIndex)
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("create obs bill txn failed, err %s", err.Error())
	}

	return nil
}

func (act SyncAction) getHuaweiMainAccount(kt *kit.Kit, mainAccountID string) (
	*accountsetproto.MainAccountGetResult[accountsetcore.HuaWeiMainAccountExtension], error) {

	mainAccount, err := actcli.GetDataService().HuaWei.MainAccount.Get(kt, mainAccountID)
	if err != nil {
		logs.Warnf("get main account by id %s failed, err %s, rid: %s", mainAccountID, err.Error(), kt.Rid)
		return nil, err
	}
	return mainAccount, nil
}

func (act SyncAction) convertHuaweiBill(
	kt *kit.Kit, syncOpt *SyncOption, result *databill.HuaweiBillItemListResult, setIndex string,
	mainAccount *accountsetproto.MainAccountGetResult[accountsetcore.HuaWeiMainAccountExtension]) (
	[]*tableobs.OBSBillItemHuawei, error) {

	item := result.Details[0]
	currency := item.Currency
	if len(currency) == 0 {
		logs.Warnf("empty currency for item %v, rid: %s", item, kt.Rid)
		return nil, fmt.Errorf("empty currency for item %v, rid: %s", item, kt.Rid)
	}
	// 获取当月平均汇率
	exhangeRate, err := act.getExchangeRate(kt, currency, enumor.CurrencyRMB, syncOpt.BillYear, syncOpt.BillMonth)
	if err != nil {
		logs.Warnf("failed to get exchange rate, err %s, rid %s", err.Error(), kt.Rid)
		return nil, fmt.Errorf("failed to get exchange rate, err %s, rid %s", err.Error(), kt.Rid)
	}
	floatRate, _ := exhangeRate.Float64()
	yearM := syncOpt.BillYear*100 + syncOpt.BillMonth

	var retList []*tableobs.OBSBillItemHuawei
	for _, item := range result.Details {
		record := item.Extension.ResFeeRecordV2
		fetchTime, err := time.Parse("2006-01-02T15:04:05Z07:00", item.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed, to parse time %s, err %s", item.CreatedAt, err.Error())
		}
		newItem := &tableobs.OBSBillItemHuawei{
			SetIndex:                  setIndex,
			Vendor:                    string(syncOpt.Vendor),
			MainAccountID:             syncOpt.MainAccountID,
			BillYear:                  int64(syncOpt.BillYear),
			BillMonth:                 int64(syncOpt.BillMonth),
			EffectiveTime:             converter.PtrToVal[string](record.EffectiveTime),
			ExpireTime:                converter.PtrToVal[string](record.ExpireTime),
			ProductID:                 converter.PtrToVal[string](record.ProductId),
			ProductName:               converter.PtrToVal[string](record.ProductName),
			OrderID:                   converter.PtrToVal[string](record.OrderId),
			Amount:                    fmt.Sprintf("%f", converter.PtrToVal[float64](record.Amount)),
			MeasureID:                 fmt.Sprintf("%d", converter.PtrToVal[int32](record.MeasureId)),
			UsageType:                 converter.PtrToVal[string](record.UsageType),
			Usages:                    fmt.Sprintf("%f", converter.PtrToVal[float64](record.Usage)),
			UsageMeasureID:            fmt.Sprintf("%d", converter.PtrToVal[int32](record.UsageMeasureId)),
			FreeResourceUsage:         fmt.Sprintf("%f", converter.PtrToVal[float64](record.FreeResourceUsage)),
			FreeResourceMeasureID:     fmt.Sprintf("%d", converter.PtrToVal[int32](record.FreeResourceMeasureId)),
			CloudServiceType:          converter.PtrToVal[string](record.CloudServiceType),
			Region:                    converter.PtrToVal[string](record.Region),
			ResourceType:              converter.PtrToVal[string](record.ResourceType),
			ChargeMode:                converter.PtrToVal[string](record.ChargeMode),
			ResourceTag:               converter.PtrToVal[string](record.ResourceTag),
			ResourceName:              converter.PtrToVal[string](record.ResourceName),
			ResourceID:                converter.PtrToVal[string](record.ResourceId),
			BillType:                  fmt.Sprintf("%d", converter.PtrToVal[int32](record.BillType)),
			EnterpriseProjectID:       converter.PtrToVal[string](record.EnterpriseProjectId),
			PeriodType:                converter.PtrToVal[string](record.PeriodType),
			Spot:                      "",
			RiUsage:                   fmt.Sprintf("%f", converter.PtrToVal[float64](record.RiUsage)),
			RiUsageMeasureID:          fmt.Sprintf("%d", converter.PtrToVal[int32](record.RiUsageMeasureId)),
			OfficialAmount:            fmt.Sprintf("%f", converter.PtrToVal[float64](record.OfficialAmount)),
			DiscountAmount:            fmt.Sprintf("%f", converter.PtrToVal[float64](record.DiscountAmount)),
			CashAmount:                fmt.Sprintf("%f", converter.PtrToVal[float64](record.CashAmount)),
			CreditAmount:              fmt.Sprintf("%f", converter.PtrToVal[float64](record.CreditAmount)),
			CouponAmount:              fmt.Sprintf("%f", converter.PtrToVal[float64](record.CouponAmount)),
			FlexipurchaseCouponAmount: fmt.Sprintf("%f", converter.PtrToVal[float64](record.FlexipurchaseCouponAmount)),
			StoredCardAmount:          fmt.Sprintf("%f", converter.PtrToVal[float64](record.StoredCardAmount)),
			BonusAmount:               fmt.Sprintf("%f", converter.PtrToVal[float64](record.BonusAmount)),
			DebtAmount:                fmt.Sprintf("%f", converter.PtrToVal[float64](record.DebtAmount)),
			AdjustmentAmount:          fmt.Sprintf("%f", converter.PtrToVal[float64](record.AdjustmentAmount)),
			SpecSize:                  fmt.Sprintf("%f", converter.PtrToVal[float64](record.SpecSize)),
			SpecSizeMeasureID:         fmt.Sprintf("%d", converter.PtrToVal[int32](record.SpecSizeMeasureId)),
			AccountName:               mainAccount.Extension.CloudMainAccountName,
			AccountType:               string(mainAccount.BusinessType),
			ProductId:                 int32(mainAccount.OpProductID),
			YearMonth:                 int32(yearM),
			FetchTime:                 fetchTime.Format("2006-01-02 15:04:05"),
			TotalCount:                int32(len(result.Details)),
			Rate:                      floatRate,
			RealCost:                  &types.Decimal{Decimal: item.Cost.Mul(decimal.NewFromFloat(floatRate))},
		}
		retList = append(retList, newItem)
	}
	return retList, nil
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
