/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
	asproto "hcm/pkg/api/data-service/account-set"
	billproto "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tableobs "hcm/pkg/dal/table/obs"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

func (act SyncAction) getAwsMainAccount(kt *kit.Kit, mainAccountID string) (
	*asproto.MainAccountGetResult[accountsetcore.AwsMainAccountExtension], error) {

	mainAccount, err := actcli.GetDataService().Aws.MainAccount.Get(kt, mainAccountID)
	if err != nil {
		logs.Warnf("get aws main account by id %s failed, err %s, rid: %s", mainAccountID, err.Error(), kt.Rid)
		return nil, err
	}
	return mainAccount, nil
}
func (act SyncAction) doBatchSyncAwsBillitem(kt *kit.Kit,
	mainAccount *asproto.MainAccountGetResult[accountsetcore.AwsMainAccountExtension],
	syncOpt *SyncOption) error {
	for start := syncOpt.Start; start < syncOpt.Start+syncOpt.Limit; start = start + uint64(core.DefaultMaxPageLimit) {
		var err error
		for retry := 0; retry < defaultRetryTimes; retry++ {
			if err = act.doSyncAwsBillItem(
				kt, mainAccount, syncOpt, start, uint64(core.DefaultMaxPageLimit)); err != nil {

				logs.Warnf("do sync aws bill %v, start %d, limit %d, retry %d, rid %s",
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

func (act SyncAction) doSyncAwsBillItem(kt *kit.Kit,
	mainAccount *asproto.MainAccountGetResult[accountsetcore.AwsMainAccountExtension],
	syncOpt *SyncOption, start, limit uint64) error {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("vendor", syncOpt.Vendor),
		tools.RuleEqual("bill_year", syncOpt.BillYear),
		tools.RuleEqual("bill_month", syncOpt.BillMonth),
		tools.RuleEqual("main_account_id", syncOpt.MainAccountID),
	}
	listFilter := tools.ExpressionAnd(expressions...)

	// 获取分账后的bill item
	result, err := actcli.GetDataService().Aws.Bill.ListBillItem(kt, &core.ListReq{
		Filter: listFilter,
		Page:   &core.BasePage{Start: uint32(start), Limit: uint(limit)}})
	if err != nil {
		logs.Warnf("list aws bill item by option %v failed, err %s, rid: %s", syncOpt, err.Error(), kt.Rid)
		return err
	}

	if len(result.Details) == 0 {
		logs.Infof("get no bill item for main_account_id %s %d-%d %d-%d, rid: %s",
			syncOpt.MainAccountID, syncOpt.BillYear, syncOpt.BillMonth, start, limit, kt.Rid)
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
		if err := actcli.GetObsDaoSet().OBSBillItemAws().DeleteWithTx(
			kt, txn, deleteFilter, uint64(core.DefaultMaxPageLimit)); err != nil {
			logs.Warnf("delete aws obs bill item by filter %v failed, err %s, rid: %s",
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
	finalItems, err := act.convertAwsBill(kt, syncOpt, result, setIndex, mainAccount)
	if err != nil {
		logs.Warnf("convert obs aws bill failed, err %s, rid: %s", err.Error(), kt.Rid)
		return err
	}
	_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if _, err := actcli.GetObsDaoSet().OBSBillItemAws().CreateWithTx(kt, txn, finalItems); err != nil {
			logs.Warnf("delete aws obs bill item by filter %s failed, err %s, rid: %s",
				deleteFilter, err.Error(), kt.Rid)
			return nil, err
		}
		logs.Infof("create obs aws bill for %s successfully", setIndex)
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("create obs bill txn failed, err %s", err.Error())
	}

	return nil
}

func (act SyncAction) convertAwsBill(kt *kit.Kit, syncOpt *SyncOption, result *billproto.AwsBillItemListResult,
	setIndex string, mainAccount *asproto.MainAccountGetResult[accountsetcore.AwsMainAccountExtension]) (
	[]*tableobs.OBSBillItemAws, error) {

	yearM := syncOpt.BillYear*100 + syncOpt.BillMonth
	item := result.Details[0]
	currency := item.Currency
	if len(currency) == 0 {
		logs.Warnf("empty currency for item %v, rid: %s", item, kt.Rid)
		return nil, fmt.Errorf("empty currency for item %v, rid: %s", item, kt.Rid)
	}
	// 获取当月平均汇率
	exchangeRate, err := act.getExchangeRate(kt, currency, enumor.CurrencyRMB, syncOpt.BillYear, syncOpt.BillMonth)
	if err != nil {
		logs.Warnf("failed to get exchange rate, err %s, rid %s", err.Error(), kt.Rid)
		return nil, fmt.Errorf("failed to get exchange rate, err %s, rid %s", err.Error(), kt.Rid)
	}
	floatRate, _ := exchangeRate.Float64()

	// 1 国内 2 国际
	var regionCode = int32(2)
	if mainAccount.Site == enumor.MainAccountChinaSite {
		regionCode = 1
	}

	var retList = make([]*tableobs.OBSBillItemAws, 0, len(result.Details))
	for _, item := range result.Details {
		record := item.Extension

		rawCost, err := decimal.NewFromString(record.LineItemNetUnblendedCost)
		if err != nil {
			logs.Errorf("fail create decimal cost (%s) from record, err %s, rid: %s",
				record.LineItemNetUnblendedCost, err.Error(), kt.Rid)
		}
		newItem := &tableobs.OBSBillItemAws{
			SetIndex:          setIndex,
			MainAccountID:     syncOpt.MainAccountID,
			BillYear:          int64(syncOpt.BillYear),
			BillMonth:         int64(syncOpt.BillMonth),
			Vendor:            string(syncOpt.Vendor),
			YearMonth:         int32(yearM),
			Rate:              floatRate,
			Cost:              &types.Decimal{Decimal: rawCost.Mul(*exchangeRate)},
			ProductID:         int32(mainAccount.OpProductID),
			LinkedAccountName: mainAccount.Extension.CloudMainAccountName,
			Region:            regionCode,
			Memo:              "",

			BillPayerAccountID:                     record.BillPayerAccountId,
			LineItemUsageAccountID:                 record.LineItemUsageAccountId,
			BillInvoiceID:                          record.BillInvoiceId,
			BillBillingEntity:                      record.BillBillingEntity,
			LineItemProductCode:                    record.LineItemProductCode,
			ProductProductFamily:                   record.ProductProductFamily,
			ProductProductName:                     record.ProductProductName,
			LineItemUsageType:                      record.LineItemUsageType,
			ProductInstanceType:                    record.ProductInstanceType,
			ProductRegion:                          record.ProductRegion,
			ProductLocation:                        record.ProductLocation,
			LineItemResourceID:                     record.LineItemResourceId,
			PricingTerm:                            record.PricingTerm,
			LineItemLineItemType:                   record.LineItemLineItemType,
			LineItemLineItemDescription:            record.LineItemLineItemDescription,
			LineItemUsageStartDate:                 record.LineItemUsageStartDate,
			LineItemUsageEndDate:                   record.LineItemUsageEndDate,
			LineItemUsageAmount:                    record.LineItemUsageAmount,
			PricingUnit:                            record.PricingUnit,
			PricingPublicOnDemandRate:              record.PricingPublicOnDemandRate,
			LineItemUnblendedRate:                  record.LineItemUnblendedRate,
			LineItemNetUnblendedRate:               record.LineItemNetUnblendedRate,
			SavingsPlanSavingsPlanRate:             record.SavingsPlanSavingsPlanRate,
			PricingPublicOnDemandCost:              record.PricingPublicOnDemandCost,
			LineItemUnblendedCost:                  record.LineItemUnblendedCost,
			LineItemNetUnblendedCost:               record.LineItemNetUnblendedCost,
			SavingsPlanSavingsPlanEffectiveCost:    record.SavingsPlanNetSavingsPlanEffectiveCost,
			SavingsPlanSavingsPlanNetEffectiveCost: record.SavingsPlanSavingsPlanEffectiveCost,
			ReservationEffectiveCost:               record.ReservationNetEffectiveCost,
			ReservationNetEffectiveCost:            record.ReservationEffectiveCost,
			LineItemCurrencyCode:                   record.LineItemCurrencyCode,
			LineItemOperation:                      record.LineItemOperation,
			DiscountPrivateRateDiscount:            record.DiscountPrivateRateDiscount,
			DiscountEDPDiscount:                    record.DiscountEdpDiscount,
		}
		retList = append(retList, newItem)
	}
	return retList, nil
}
