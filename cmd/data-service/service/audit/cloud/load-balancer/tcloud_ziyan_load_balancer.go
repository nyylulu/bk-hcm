package loadbalancer

import (
	"fmt"

	"hcm/pkg/api/core"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableaudit "hcm/pkg/dal/table/audit"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

func (c *LoadBalancer) tcloudZiyanUrlRuleUpdateAuditBuild(kt *kit.Kit, lbl tablelb.LoadBalancerListenerTable,
	updates []protoaudit.CloudResourceUpdateInfo) ([]*tableaudit.AuditTable, error) {

	ids := slice.Map(updates, func(u protoaudit.CloudResourceUpdateInfo) string { return u.ResID })

	idListenerRuleMap, err := listTCloudZiyanUrlRule(kt, c.dao, lbl.ID, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(updates))
	for _, one := range updates {
		rule, exist := idListenerRuleMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: lbl.CloudID,
			ResName:    lbl.Name,
			ResType:    enumor.ListenerAuditResType,
			Action:     enumor.Update,
			BkBizID:    lbl.BkBizID,
			Vendor:     lbl.Vendor,
			AccountID:  lbl.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.ChildResAuditData{
					ChildResType: enumor.UrlRuleAuditResType,
					Action:       enumor.Update,
					ChildRes:     rule,
				},
				Changed: one.UpdateFields,
			},
		})
	}

	return audits, nil

}

func (c *LoadBalancer) tcloudZiyanUrlRuleDeleteAuditBuild(kt *kit.Kit, lbl tablelb.LoadBalancerListenerTable,
	deletes []protoaudit.CloudResourceDeleteInfo) ([]*tableaudit.AuditTable, error) {

	ids := slice.Map(deletes, func(u protoaudit.CloudResourceDeleteInfo) string { return u.ResID })

	idRuleMap, err := listTCloudZiyanUrlRule(kt, c.dao, lbl.ID, ids)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		ruleInfo, exist := idRuleMap[one.ResID]
		if !exist {
			continue
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: ruleInfo.CloudID,
			ResName:    ruleInfo.Name,
			ResType:    enumor.UrlRuleAuditResType,
			Action:     enumor.Delete,
			BkBizID:    lbl.BkBizID,
			Vendor:     lbl.Vendor,
			AccountID:  lbl.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: ruleInfo,
			},
		})
	}

	return audits, nil
}

func (c *LoadBalancer) tcloudZiyanUrlRuleDeleteByDomainAuditBuild(kt *kit.Kit, lbl tablelb.LoadBalancerListenerTable,
	deletes []protoaudit.CloudResourceDeleteInfo) ([]*tableaudit.AuditTable, error) {

	domains := slice.Map(deletes, func(u protoaudit.CloudResourceDeleteInfo) string { return u.ResID })

	domainRuleMap, err := listTCloudZiyanUrlRuleByDomain(kt, c.dao, lbl.ID, domains)
	if err != nil {
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(deletes))
	for _, one := range deletes {
		rules, exist := domainRuleMap[one.ResID]
		if !exist {
			// 找不到与域名，返回错误
			return nil, fmt.Errorf("fail to find rule while delete url by domain: %s", one.ResID)
		}
		// add domain and each into audits
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ResID,
			CloudResID: one.ResID,
			ResName:    one.ResID,
			ResType:    enumor.UrlRuleDomainAuditResType,
			Action:     enumor.Delete,
			BkBizID:    lbl.BkBizID,
			Vendor:     lbl.Vendor,
			AccountID:  lbl.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: one.ResID,
			},
		})
		for _, rule := range rules {
			audits = append(audits, &tableaudit.AuditTable{
				ResID:      rule.ID,
				CloudResID: rule.CloudID,
				ResName:    rule.Name,
				ResType:    enumor.UrlRuleAuditResType,
				Action:     enumor.Delete,
				BkBizID:    lbl.BkBizID,
				Vendor:     lbl.Vendor,
				AccountID:  lbl.AccountID,
				Operator:   kt.User,
				Source:     kt.GetRequestSource(),
				Rid:        kt.Rid,
				AppCode:    kt.AppCode,
				Detail: &tableaudit.BasicDetail{
					Data: rule,
				},
			})
		}
	}

	return audits, nil
}

func listTCloudZiyanUrlRule(kt *kit.Kit, dao dao.Set, lblID string,
	ruleIds []string) (map[string]tablelb.TCloudZiyanLbUrlRuleTable, error) {

	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(tools.RuleEqual("lbl_id", lblID), tools.RuleIn("id", ruleIds)),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancerTCloudZiyanUrlRule().List(kt, opt)
	if err != nil {
		logs.Errorf("list tcloud-ziyan url rule of  listener(id=%s) failed, err: %v, ids: %v, rid: %s",
			lblID, err, ruleIds, kt.Rid)
		return nil, err
	}

	result := make(map[string]tablelb.TCloudZiyanLbUrlRuleTable, len(list.Details))
	for _, one := range list.Details {
		result[one.ID] = *one
	}

	return result, nil
}

func listTCloudZiyanUrlRuleByDomain(kt *kit.Kit, dao dao.Set, lblID string,
	domains []string) (map[string][]tablelb.TCloudZiyanLbUrlRuleTable, error) {

	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(tools.RuleEqual("lbl_id", lblID), tools.RuleIn("domain", domains)),
		Page:   core.NewDefaultBasePage(),
	}
	list, err := dao.LoadBalancerTCloudZiyanUrlRule().List(kt, opt)
	if err != nil {
		logs.Errorf("list tcloud-ziyan url rule of listener(id=%s) failed, err: %v, domains: %v, rid: %s",
			lblID, err, domains, kt.Rid)
		return nil, err
	}

	result := make(map[string][]tablelb.TCloudZiyanLbUrlRuleTable, len(list.Details))
	for _, one := range list.Details {
		result[one.Domain] = append(result[one.Domain], *one)
	}

	return result, nil
}
