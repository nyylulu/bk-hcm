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

package savingsplans

import (
	"errors"
	"fmt"
	"net/http"

	"hcm/cmd/account-server/logics/audit"
	"hcm/cmd/account-server/service/capability"
	typesBill "hcm/pkg/adaptor/types/bill"
	asbillapi "hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	accountcore "hcm/pkg/api/core/account-set"
	"hcm/pkg/api/hc-service/bill"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/auth"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/shopspring/decimal"
)

// InitService initial the main account service
func InitService(c *capability.Capability) {
	svc := &service{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("AwsQuerySavingsPlanSavedCost", http.MethodPost,
		"/vendors/aws/savings_plans/saved_cost/query", svc.AwsQuerySavingsPlanSavedCost)
	h.Load(c.WebService)
}

type service struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}

// AwsQuerySavingsPlanSavedCost query aws saving plans
func (s *service) AwsQuerySavingsPlanSavedCost(cts *rest.Contexts) (any, error) {
	req := new(asbillapi.AwsSPSavedCostReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// TODO 暴露账号接口后改为账号鉴权并设置RootAccountID 为必填
	authReq := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AwsSavingsPlansCost, Action: meta.Find}}
	err := s.authorizer.AuthorizeWithPerm(cts.Kit, authReq)
	if err != nil {
		return nil, err
	}
	var rootAccountCloudID string
	if req.RootAccountID != "" {
		// 尝试查询用户请求中携带的根账号
		info, err := s.client.DataService().Global.RootAccount.GetBasicInfo(cts.Kit, req.RootAccountID)
		if err != nil {
			logs.Errorf("fail to get root account info, err: %v, root id: %s, rid: %s",
				err, req.RootAccountID, cts.Kit.Rid)
			return nil, err
		}
		rootAccountCloudID = info.CloudID
	}

	// 读取 aws SavingPlans 配置，获取指定的sp arn前缀
	spOpt, err := s.matchingSavingsPlanOption(rootAccountCloudID)
	if err != nil {
		logs.Errorf("matching option failed by root account cloud id %s, err: %v, rid: %s",
			rootAccountCloudID, err, cts.Kit.Rid)
		return nil, fmt.Errorf("matching option failed, err: %v", err)
	}
	// 因为支持不输入根账号，因此需要再查一次根账号
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("cloud_id", spOpt.RootAccountCloudID),
			tools.RuleEqual("vendor", enumor.Aws)),
		Page: core.NewDefaultBasePage(),
	}
	rootResp, err := s.client.DataService().Global.RootAccount.List(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("fail to list root account, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(rootResp.Details) == 0 {
		logs.Errorf("root account not found by cloud id %s,  rid: %s", spOpt.RootAccountCloudID, cts.Kit.Rid)
		return nil, errors.New("root account not found")
	}
	rootAccount := rootResp.Details[0]

	// 依据productIDs筛选AccountCloudIDs
	usageAccountIDs, err := s.filterAccountCloudIDs(cts.Kit, req)
	if err != nil {
		logs.Errorf("filter account failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return s.querySpInfo(cts.Kit, rootAccount.ID, spOpt, usageAccountIDs, req)
}

func (s *service) querySpInfo(kt *kit.Kit, rootID string, spOpt *cc.AwsSavingsPlansOption, usageAccountIDs []string,
	req *asbillapi.AwsSPSavedCostReq) (*asbillapi.AwsSPCostResult, error) {

	querySPCostReq := &bill.QueryAwsSavingsPlanCostReq{
		RootAccountID:   rootID,
		SavingPlansArn:  spOpt.SpArnPrefix,
		UsageAccountIDs: usageAccountIDs,
		Year:            req.Year,
		Month:           req.Month,
		StartDay:        req.StartDay,
		EndDay:          req.EndDay,
		Page:            req.Page,
	}
	spCostResp, err := s.client.HCService().Aws.Bill.QuerySavingsPlanCostList(kt, querySPCostReq)
	if err != nil {
		logs.Errorf("list sp saved cost failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list sp saved cost failed, err: %v", err)
	}
	if spCostResp.Count > 0 {
		return &asbillapi.AwsSPCostResult{Count: spCostResp.Count}, nil
	}
	mainCloudIdMap, err := s.getMainAccountMap(kt, spOpt.SpPurchaseAccountCloudID, spCostResp.Details)
	if err != nil {
		logs.Errorf("fail to get main account for sp cost, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	spAccount := mainCloudIdMap[spOpt.SpPurchaseAccountCloudID]
	spInfoList := make([]asbillapi.AwsAccountSPCost, 0, len(spCostResp.Details))
	batchTotal := decimal.Zero
	for i := range spCostResp.Details {
		spCost := spCostResp.Details[i]
		mainAccount := mainCloudIdMap[spCost.AccountCloudID]
		spInfoList = append(spInfoList, convSpCostItem(mainAccount, spCost, spAccount))
		savedCost, err := decimal.NewFromString(spCost.SPSavedCost)
		if err != nil {
			logs.Errorf("fail to convert sp saved cost to decimal, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		batchTotal = batchTotal.Add(savedCost)
	}
	result := &asbillapi.AwsSPCostResult{
		Details:    spInfoList,
		BatchTotal: batchTotal.String(),
	}
	return result, nil
}

func convSpCostItem(mainAccount *accountcore.BaseMainAccount, spCost typesBill.AwsSavingsPlansCost,
	spAccount *accountcore.BaseMainAccount) asbillapi.AwsAccountSPCost {
	return asbillapi.AwsAccountSPCost{
		// 暂不暴露账号信息
		// MainAccountID:          mainAccount.ID,
		// MainAccountCloudID:     mainAccount.CloudID,
		// MainAccountManagers:    mainAccount.Managers,
		// MainAccountBakManagers: mainAccount.BakManagers,
		ProductId:       mainAccount.OpProductID,
		SpArn:           spCost.SpArn,
		SpManagers:      spAccount.Managers,
		SpBakManagers:   spAccount.BakManagers,
		UnblendedCost:   spCost.UnblendedCost,
		SPEffectiveCost: spCost.SPEffectiveCost,
		SPNetCost:       spCost.SPNetCost,
		SPSavedCost:     spCost.SPSavedCost,
	}
}

func (s *service) getMainAccountMap(kt *kit.Kit, spCloudId string, spCostList []typesBill.AwsSavingsPlansCost) (
	map[string]*accountcore.BaseMainAccount, error) {

	// 查询包含 sp purchase account 在内的main account
	mainAccountCloudIDs := map[string]struct{}{spCloudId: {}}
	// 补全账号信息
	for i := range spCostList {
		spCost := spCostList[i]
		mainAccountCloudIDs[spCost.AccountCloudID] = struct{}{}
	}
	mainCloudIdMap := make(map[string]*accountcore.BaseMainAccount, len(mainAccountCloudIDs))
	// 查询账号信息
	for _, cloudIds := range slice.Split(cvt.MapKeyToSlice(mainAccountCloudIDs), constant.BatchOperationMaxLimit) {
		listMainAccountReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", enumor.Aws),
				tools.RuleIn("cloud_id", cloudIds),
			),
			Page: core.NewDefaultBasePage(),
		}
		accResp, err := s.client.DataService().Global.MainAccount.List(kt, listMainAccountReq)
		if err != nil {
			logs.Errorf("fail to list aws main account, err: %v, cloud ids: %v, rid: %s", err, cloudIds, kt.Rid)
			return nil, err
		}
		if len(cloudIds) != len(accResp.Details) {
			gotCloudIds := slice.Map(accResp.Details, func(a *accountcore.BaseMainAccount) string { return a.CloudID })
			notFoundIds := slice.NotIn(cloudIds, gotCloudIds)
			logs.Errorf("some account can not be found: %v, rid: %s", notFoundIds, kt.Rid)
			return nil, fmt.Errorf("some account can not be found: %v", notFoundIds)
		}
		for i := range accResp.Details {
			mainCloudIdMap[accResp.Details[i].CloudID] = accResp.Details[i]
		}
	}
	return mainCloudIdMap, nil
}

func (s *service) matchingSavingsPlanOption(rootAccountCloud string) (*cc.AwsSavingsPlansOption, error) {

	optList := cc.AccountServer().BillAllocation.AwsSavingsPlans
	if len(rootAccountCloud) == 0 && len(optList) != 0 {
		// 暂时支持不输入根账号，直接返回第一个配置
		return cvt.ValToPtr(optList[0]), nil
	}
	for i := range optList {
		spOpt := optList[i]
		if spOpt.RootAccountCloudID == rootAccountCloud {
			return cvt.ValToPtr(spOpt), nil
		}
	}
	return nil, fmt.Errorf("no aws savings plans configuration found for root account: %s", rootAccountCloud)
}

func (s *service) filterAccountCloudIDs(kit *kit.Kit, req *asbillapi.AwsSPSavedCostReq) ([]string, error) {

	usageAccountIDs := make([]string, 0)
	if len(req.MainAccountIDs) == 0 && len(req.MainAccountCloudIDs) == 0 && len(req.ProductIDs) == 0 {
		return usageAccountIDs, nil
	}
	rules := make([]filter.RuleFactory, 0, 2)
	if len(req.MainAccountIDs) != 0 {
		rules = append(rules, tools.RuleIn("id", req.MainAccountIDs))
	}
	if len(req.MainAccountCloudIDs) != 0 {
		rules = append(rules, tools.RuleIn("cloud_id", req.MainAccountCloudIDs))
	}
	if len(req.ProductIDs) != 0 {
		rules = append(rules, tools.RuleIn("op_product_id", req.ProductIDs))
	}

	listMainAccReq := &core.ListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: rules},
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"cloud_id"},
	}

	for {
		listMainAccResult, err := s.client.DataService().Global.MainAccount.List(kit, listMainAccReq)
		if err != nil {
			logs.Errorf("list account failed, err: %v, rid: %s", err, kit.Rid)
			return nil, err
		}
		for _, mainAccItem := range listMainAccResult.Details {
			usageAccountIDs = append(usageAccountIDs, mainAccItem.CloudID)
		}

		if len(listMainAccResult.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		listMainAccReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	if len(usageAccountIDs) == 0 {
		return nil, fmt.Errorf("no account found by ids:%v, cloud ids: %v, product ids: %v",
			req.MainAccountIDs, req.MainAccountCloudIDs, req.ProductIDs)
	}

	return usageAccountIDs, nil
}
