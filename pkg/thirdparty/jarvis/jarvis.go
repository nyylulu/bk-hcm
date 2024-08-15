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

package jarvis

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	"hcm/pkg/tools/rand"
	"hcm/pkg/tools/ssl"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"
)

// Client 财务侧api
type Client interface {
	GetPeriodExchangeRate(kt *kit.Kit, req *GetPeriodExchangeRateReq) (*GetPeriodExchangeRateResp, error)
}

// jarvisCli ...
type jarvisCli struct {
	cfg *cc.Jarvis
	cli rest.ClientInterface
}

// NewJarvis ...
func NewJarvis(cfg *cc.Jarvis, reg prometheus.Registerer) (Client, error) {

	tls := &ssl.TLSConfig{
		InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
		CertFile:           cfg.TLS.CertFile,
		KeyFile:            cfg.TLS.KeyFile,
		CAFile:             cfg.TLS.CAFile,
		Password:           cfg.TLS.Password,
	}
	cli, err := client.NewClient(tls)
	if err != nil {
		return nil, err
	}
	c := &client.Capability{
		Client: cli,
		Discover: &discovery{
			servers: cfg.Endpoints,
		},
		MetricOpts: client.MetricOption{Register: reg},
	}
	restCli := rest.NewClient(c, "/")

	return &jarvisCli{cfg: cfg, cli: restCli}, nil
}

// GetPeriodExchangeRate ...
func (j *jarvisCli) GetPeriodExchangeRate(kt *kit.Kit, req *GetPeriodExchangeRateReq) (
	*GetPeriodExchangeRateResp, error) {

	timestamp := time.Now().Unix()
	nonce := rand.String(8)
	periodName := req.StartPeriod.String() + "," + req.EndPeriod.String()
	params := map[string]string{
		"appid":       j.cfg.AppID,
		"timestamp":   strconv.FormatInt(timestamp, 10),
		"nonce":       nonce,
		"pageIndex":   strconv.FormatInt(req.PageIndex, 10),
		"pageSize":    strconv.FormatInt(req.PageSize, 10),
		"period_name": periodName,
	}
	if len(req.UserConversionType) > 0 {
		params["userConversionType"] = req.UserConversionType
	}
	if len(req.ToCurrency) > 0 {
		params["to_Currency"] = string(req.ToCurrency)
	}
	if len(req.FromCurrency) > 0 {
		params["from_Currency"] = string(req.FromCurrency)
	}

	signature, err := j.getSignature(params)
	if err != nil {
		logs.Errorf("fail to calculate signature, err: %v, params: %v, rid: %s", params, err, kt.Rid)
		return nil, err
	}

	resp := new(GetPeriodExchangeRateResp)
	err = j.cli.Get().
		WithContext(kt.Ctx).
		WithHeaders(kt.Header()).
		WithParams(params).
		WithParam("signature", signature).
		SubResourcef("/api/EBSGLDSService/1/getPeriodExchangeRate").
		Do().
		Into(resp)
	if err != nil {
		logs.Errorf("fail to call jarvisCli api getPeriodExchangeRate, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp, nil
}

func (j *jarvisCli) getSignature(params map[string]string) (string, error) {

	var dataParams string
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接
	for _, k := range keys {
		dataParams = dataParams + k + "=" + url.QueryEscape(params[k]) + "&"
	}

	stringSignTmp := dataParams[0 : len(dataParams)-1]
	signParam := stringSignTmp + "&key=" + j.cfg.AppKey

	h := hmac.New(sha256.New, []byte(j.cfg.AppKey))
	_, err := h.Write([]byte(signParam))
	if err != nil {
		log.Printf("hmac sha256 failed: %s\n", err.Error())
		return "", err
	}
	sha := hex.EncodeToString(h.Sum(nil))
	signature := strings.ToUpper(sha)
	return signature, nil
}

// PeriodExchangeRate ...
type PeriodExchangeRate struct {
	FromCurrency          enumor.CurrencyCode `json:"fromCurrency"`
	ToCurrency            enumor.CurrencyCode `json:"toCurrency"`
	PeriodName            string              `json:"periodName"`
	UserConversionType    string              `json:"userConversionType"`
	ConversionRate        decimal.Decimal     `json:"conversionRate"`
	InverseConversionRate decimal.Decimal     `json:"inverseConversionRate"`
}

// Response ...
type Response[T any] struct {
	TotalRecords int `json:"totalRecords"`
	Data         []T `json:"data"`
}

// GetPeriodExchangeRateResp ...
type GetPeriodExchangeRateResp = Response[PeriodExchangeRate]

// YearMonth ...
type YearMonth struct {
	Year  int `json:"year" validate:"required"`
	Month int `json:"month" validate:"required,min=1,max=12"`
}

// Validate ...
func (ym *YearMonth) Validate() error {
	return validator.Validate.Struct(ym)
}

func (ym *YearMonth) String() string {
	return fmt.Sprintf("%d-%02d", ym.Year, ym.Month)
}

// GetPeriodExchangeRateReq ...
type GetPeriodExchangeRateReq struct {
	// 原货币
	FromCurrency enumor.CurrencyCode `json:"from_Currency" `
	// 目标货币
	ToCurrency  enumor.CurrencyCode `json:"to_Currency"`
	StartPeriod YearMonth           `json:"start_period" validate:"required"`
	EndPeriod   YearMonth           `json:"end_period" validate:"required"`
	// 汇率类型
	UserConversionType string `json:"userConversionType"`
	// 从1开始
	PageIndex int64 `json:"pageIndex" validate:"required,min=1"`
	PageSize  int64 `json:"pageSize" validate:"required,min=1,max=500"`
}

// Validate ...
func (r *GetPeriodExchangeRateReq) Validate() error {
	return validator.Validate.Struct(r)
}

const (

	// ConversionTypeCorporate Corporate
	ConversionTypeCorporate = "Corporate"
	// ConversionTypePeriodAverage 周期平均
	ConversionTypePeriodAverage = "Period Average"
	// ConversionTypePeriodEnd Period End
	ConversionTypePeriodEnd = "Period End"
	// ConversionTypeOverseas Overseas
	ConversionTypeOverseas = "Overseas"
	// ConversionTypeOverseasEnd OverseasEnd
	ConversionTypeOverseasEnd = "Overseas End"
)
