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

package bill

import (
	"errors"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/bill"
	databill "hcm/pkg/api/data-service/bill"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	"hcm/pkg/thirdparty/jarvis"
	cvt "hcm/pkg/tools/converter"
)

// ExchangeRateOpt ...
type ExchangeRateOpt struct {
	Sd      serviced.ServiceDiscover
	Jarvis  jarvis.Client
	DataCli *dataservice.Client
	// 留空表示同步所有来源币种
	FromCurrencyCodes []enumor.CurrencyCode `json:"fromCurrency"`
	// 目标币种，不支持留空
	ToCurrencyCodes []enumor.CurrencyCode `json:"toCurrency"`
	// 循环间隔
	LoopInterval time.Duration
}

// ExchangeRateController ...
type ExchangeRateController struct {
	sd           serviced.ServiceDiscover
	jarvis       jarvis.Client
	dataCli      *dataservice.Client
	fromCurrency []enumor.CurrencyCode
	toCurrency   []enumor.CurrencyCode
	kt           *kit.Kit
	loopInterval time.Duration
}

// NewExchangeRateController ...
func NewExchangeRateController(opt *ExchangeRateOpt) (*ExchangeRateController, error) {
	if opt.Jarvis == nil {
		return nil, errors.New("jarvis is required")
	}
	if opt.DataCli == nil {
		return nil, errors.New("data service client is required")
	}

	if len(opt.ToCurrencyCodes) == 0 {
		return nil, errors.New("to currency code is required")
	}

	return &ExchangeRateController{
		jarvis:       opt.Jarvis,
		dataCli:      opt.DataCli,
		fromCurrency: opt.FromCurrencyCodes,
		toCurrency:   opt.ToCurrencyCodes,
		sd:           opt.Sd,
	}, nil
}

// Run ...
func (c *ExchangeRateController) Run() {
	kt := getInternalKit()
	c.loop(kt)

}
func (c *ExchangeRateController) loop(kt *kit.Kit) {

	time.Sleep(10 * time.Second)
	if c.sd.IsMaster() {
		sub := kt.NewSubKit()
		c.sync(sub)
	}
	tick := time.Tick(c.loopInterval)
	for {
		select {
		case <-tick:
			if c.sd.IsMaster() {
				sub := kt.NewSubKit()
				c.sync(sub)
			}
		case <-kt.Ctx.Done():
			logs.Infof("exchange rate controller context done, rid: %s", kt.Rid)
			return
		}
	}
}

// syncMonth ...
func (c *ExchangeRateController) syncMonth(kt *kit.Kit, year, month int) error {
	var wantedCurrMap map[enumor.CurrencyCode]bool
	if len(c.fromCurrency) > 0 {
		wantedCurrMap = make(map[enumor.CurrencyCode]bool, len(c.fromCurrency))
		for _, c := range c.fromCurrency {
			wantedCurrMap[c] = true
		}
	}
	var addSlice []databill.ExchangeRateCreate
	for _, toCurrency := range c.toCurrency {
		dbRateMap, err := c.listRateFromDB(kt, year, month, toCurrency)
		if err != nil {
			return err
		}
		stdRateList, err := c.listRateFromJarvis(kt, year, month, toCurrency)
		if err != nil {
			return err
		}
		for _, stdRate := range stdRateList {
			if len(c.fromCurrency) > 0 && !wantedCurrMap[stdRate.FromCurrency] {
				// 跳过不需要的币种
				continue
			}

			// 不存在则创建汇率
			if _, exists := dbRateMap[stdRate.FromCurrency]; !exists {
				addSlice = append(addSlice, databill.ExchangeRateCreate{
					Year:         year,
					Month:        month,
					FromCurrency: stdRate.FromCurrency,
					ToCurrency:   stdRate.ToCurrency,
					ExchangeRate: cvt.ValToPtr(stdRate.ConversionRate),
				})
			}
		}

		if len(addSlice) > 0 {
			req := &databill.BatchCreateBillExchangeRateReq{ExchangeRates: addSlice}
			_, err := c.dataCli.Global.Bill.BatchCreateExchangeRate(kt, req)
			if err != nil {
				logs.Errorf("fail to create exchange rate, err: %v, period: %d-%d, target currency: %s, rid: %s",
					err, year, month, toCurrency, kt.Rid)
				return err
			}
			logs.Infof("create exchange rate success, count: %d, period: %d-%d, target currency: %s, rid: %s",
				len(addSlice), year, month, toCurrency, kt.Rid)
		}

	}
	return nil
}

func (c *ExchangeRateController) listRateFromJarvis(kt *kit.Kit, year int, month int, toCurrency enumor.CurrencyCode) (
	[]jarvis.PeriodExchangeRate, error) {

	jReq := &jarvis.GetPeriodExchangeRateReq{
		// 暂不指定原货币，仅以目标货币筛选汇率
		// FromCurrency: fromCurrency,
		ToCurrency:  toCurrency,
		StartPeriod: jarvis.YearMonth{Month: month, Year: year},
		EndPeriod:   jarvis.YearMonth{Month: month, Year: year},
		// 仅考虑月度平均汇率
		UserConversionType: jarvis.ConversionTypePeriodAverage,
		PageIndex:          1,
		PageSize:           500,
	}
	stdRates, err := c.jarvis.GetPeriodExchangeRate(kt, jReq)
	if err != nil {
		logs.Errorf("failt get jarvis, exchange rate, err: %v, period: %d-%d, target currency: %s, rid: %s",
			err, year, month, toCurrency, kt.Rid)
		return nil, err
	}
	stdRateList := stdRates.Data
	return stdRateList, nil
}

func (c *ExchangeRateController) listRateFromDB(kt *kit.Kit, year int, month int, toCurrency enumor.CurrencyCode) (
	map[enumor.CurrencyCode]bill.ExchangeRate, error) {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("year", year),
			tools.RuleEqual("month", month),
			tools.RuleEqual("to_currency", toCurrency),
		),
		Page: core.NewDefaultBasePage(),
	}

	rateList, err := c.dataCli.Global.Bill.ListExchangeRate(kt, listReq)
	if err != nil {
		logs.Errorf("fail to list db exchange rate, err: %v, period: %d-%d, target currency: %s, rid: %s",
			err, year, month, toCurrency, kt.Rid)
		return nil, err
	}
	dbRateMap := cvt.SliceToMap(rateList.Details, func(r bill.ExchangeRate) (enumor.CurrencyCode, bill.ExchangeRate) {
		return r.FromCurrency, r
	})
	return dbRateMap, err
}

func (c *ExchangeRateController) sync(kt *kit.Kit) {

	// 同步前两个月的汇率, 当月汇率需要等待下月才有
	year, month := getLastBillMonth()
	if err := c.syncMonth(kt, year, month); err != nil {
		logs.Errorf("fail to sync exchange rate, err: %v, period: %d-%d, rid: %s",
			err, year, month, kt.Rid)
	}

	year, month = getMonthOffset(-2)
	if err := c.syncMonth(kt, year, month); err != nil {
		logs.Errorf("fail to sync exchange rate, err: %v, period: %d-%d, rid: %s",
			err, year, month, kt.Rid)
	}
}
