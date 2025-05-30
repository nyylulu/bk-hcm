package loadbalancer

import (
	"fmt"
	"net/http"
	"reflect"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/json"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

func tcloudZiyanService(h *rest.Handler, svc *lbSvc) {
	// url规则
	h.Add("BatchCreateTCloudUrlRule",
		http.MethodPost, "/vendors/tcloud-ziyan/url_rules/batch/create", svc.BatchCreateTCloudZiyanUrlRule)
	h.Add("BatchUpdateTCloudUrlRule",
		http.MethodPatch, "/vendors/tcloud-ziyan/url_rules/batch/update", svc.BatchUpdateTCloudZiyanUrlRule)
	h.Add("BatchDeleteTCloudUrlRule",
		http.MethodDelete, "/vendors/tcloud-ziyan/url_rules/batch", svc.BatchDeleteTCloudZiyanUrlRule)
	h.Add("ListTCloudUrlRule", http.MethodPost,
		"/vendors/tcloud-ziyan/load_balancers/url_rules/list", svc.ListTCloudZiyanUrlRule)
}

// ListTCloudZiyanUrlRule list tcloud url rule.
func (svc *lbSvc) ListTCloudZiyanUrlRule(cts *rest.Contexts) (any, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.LoadBalancerTCloudZiyanUrlRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud-ziyan lb url rule failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud-ziyan lb url rule failed, err: %v", err)
	}

	if req.Page.Count {
		return &protocloud.TCloudURLRuleListResult{Count: result.Count}, nil
	}

	details := make([]corelb.TCloudLbUrlRule, 0, len(result.Details))
	for _, one := range result.Details {
		tmpOne, err := convZiyanTableToBaseTCloudLbURLRule(cts.Kit, one)
		if err != nil {
			continue
		}
		details = append(details, *tmpOne)
	}

	return &protocloud.TCloudURLRuleListResult{Details: details}, nil
}

func convZiyanTableToBaseTCloudLbURLRule(kt *kit.Kit, one *tablelb.TCloudZiyanLbUrlRuleTable) (
	*corelb.TCloudLbUrlRule, error) {

	var healthCheck *corelb.TCloudHealthCheckInfo
	err := json.UnmarshalFromString(string(one.HealthCheck), &healthCheck)
	if err != nil {
		logs.Errorf("unmarshal healthCheck failed, one: %+v, err: %v, rid: %s", one, err, kt.Rid)
		return nil, err
	}

	var certInfo *corelb.TCloudCertificateInfo
	err = json.UnmarshalFromString(string(one.Certificate), &certInfo)
	if err != nil {
		logs.Errorf("unmarshal certificate failed, one: %+v, err: %v, rid: %s", one, err, kt.Rid)
		return nil, err
	}

	return &corelb.TCloudLbUrlRule{
		ID:                 one.ID,
		CloudID:            one.CloudID,
		Name:               one.Name,
		RuleType:           one.RuleType,
		LbID:               one.LbID,
		CloudLbID:          one.CloudLbID,
		LblID:              one.LblID,
		CloudLBLID:         one.CloudLBLID,
		TargetGroupID:      one.TargetGroupID,
		CloudTargetGroupID: one.CloudTargetGroupID,
		Region:             one.Region,
		Domain:             one.Domain,
		URL:                one.URL,
		Scheduler:          one.Scheduler,
		SessionType:        one.SessionType,
		SessionExpire:      one.SessionExpire,
		HealthCheck:        healthCheck,
		Certificate:        certInfo,
		Memo:               one.Memo,
		Revision: &core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}, nil
}

// BatchDeleteTCloudZiyanUrlRule 批量删除腾讯云规则
func (svc *lbSvc) BatchDeleteTCloudZiyanUrlRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.LoadBalancerBatchDeleteReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	opt := &types.ListOption{
		Fields: []string{"id", "cloud_id"},
		Filter: req.Filter,
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dao.LoadBalancerTCloudZiyanUrlRule().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list tcloud-ziyan lb rule failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("list tcloud-ziyan lb rule failed, err: %v", err)
	}

	if len(listResp.Details) == 0 {
		return nil, nil
	}

	ruleIds := slice.Map(listResp.Details, func(one *tablelb.TCloudZiyanLbUrlRuleTable) string { return one.ID })

	_, err = svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		// 删除关联关系
		ruleFilter := tools.ContainersExpression("listener_rule_id", ruleIds)
		err := svc.dao.LoadBalancerTargetGroupListenerRuleRel().DeleteWithTx(cts.Kit, txn, ruleFilter)
		if err != nil {
			logs.Errorf("fail to delete rule target group relations, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}

		// 删除对应的规则
		delFilter := tools.ContainersExpression("id", ruleIds)
		return nil, svc.dao.LoadBalancerTCloudZiyanUrlRule().DeleteWithTx(cts.Kit, txn, delFilter)
	})
	if err != nil {
		logs.Errorf("delete rules(ids=%v) failed, err: %v, rid: %s", ruleIds, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchCreateTCloudZiyanUrlRule 批量创建腾讯云url规则 纯规则条目创建，不校验监听器， 有目标组则一起创建关联关系
func (svc *lbSvc) BatchCreateTCloudZiyanUrlRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.TCloudUrlRuleBatchCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		logs.Errorf("[ds] BatchCreateTCloudUrlRule request validate failed, err:%v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ruleModels := make([]*tablelb.TCloudZiyanLbUrlRuleTable, 0, len(req.UrlRules))
	for _, rule := range req.UrlRules {
		ruleModel, err := svc.convZiyanRule(cts.Kit, rule)
		if err != nil {
			return nil, err
		}
		ruleModels = append(ruleModels, ruleModel)
	}

	// 创建规则和关联关系
	result, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {

		ids, err := svc.dao.LoadBalancerTCloudZiyanUrlRule().BatchCreateWithTx(cts.Kit, txn, ruleModels)
		if err != nil {
			logs.Errorf("fail to batch create lb rule, err: %v, rid:%s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("batch create lb rule failed, err: %v", err)
		}
		// 根据id 创建关联关系
		relModels := make([]*tablelb.TargetGroupListenerRuleRelTable, 0, len(req.UrlRules))
		for i, rule := range req.UrlRules {
			// 跳过没有设置目标组id的规则
			if len(rule.TargetGroupID) == 0 {
				continue
			}
			// 默认设置为绑定中状态，防止同步时本地目标组rs被清掉
			relModels = append(relModels, svc.convRuleRel(cts.Kit, ids[i], rule, enumor.BindingBindingStatus))
		}
		if len(relModels) == 0 {
			return ids, nil
		}
		_, err = svc.dao.LoadBalancerTargetGroupListenerRuleRel().BatchCreateWithTx(cts.Kit, txn, relModels)
		if err != nil {
			logs.Errorf("fail to create rule rel, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create tcloud url rule but return id type is not []string, id type: %v",
			reflect.TypeOf(result).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}

func (svc *lbSvc) convZiyanRule(kt *kit.Kit, rule dataproto.TCloudUrlRuleCreate) (
	*tablelb.TCloudZiyanLbUrlRuleTable, error) {

	ruleModel := &tablelb.TCloudZiyanLbUrlRuleTable{
		CloudID:            rule.CloudID,
		Name:               rule.Name,
		RuleType:           rule.RuleType,
		LbID:               rule.LbID,
		CloudLbID:          rule.CloudLbID,
		LblID:              rule.LblID,
		CloudLBLID:         rule.CloudLBLID,
		TargetGroupID:      rule.TargetGroupID,
		CloudTargetGroupID: rule.CloudTargetGroupID,
		Region:             rule.Region,
		Domain:             rule.Domain,
		URL:                rule.URL,
		Scheduler:          rule.Scheduler,
		SessionType:        rule.SessionType,
		SessionExpire:      rule.SessionExpire,
		Memo:               rule.Memo,

		Creator: kt.User,
		Reviser: kt.User,
	}
	healthCheckJson, err := json.MarshalToString(rule.HealthCheck)
	if err != nil {
		logs.Errorf("fail to marshal health check into json, err: %v, healthcheck: %+v, rid: %s",
			err, rule.HealthCheck, kt.Rid)
		return nil, err
	}
	ruleModel.HealthCheck = tabletype.JsonField(healthCheckJson)
	certJson, err := json.MarshalToString(rule.Certificate)
	if err != nil {
		logs.Errorf("fail to marshal certificate into json, err: %v, certificate: %+v, rid: %s",
			err, rule.Certificate, kt.Rid)
		return nil, err
	}
	ruleModel.Certificate = tabletype.JsonField(certJson)
	return ruleModel, nil
}

// BatchUpdateTCloudZiyanUrlRule ..
func (svc *lbSvc) BatchUpdateTCloudZiyanUrlRule(cts *rest.Contexts) (any, error) {
	req := new(dataproto.TCloudUrlRuleBatchUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	ruleIds := slice.Map(req.UrlRules, func(one *dataproto.TCloudUrlRuleUpdate) string { return one.ID })

	healthCertMap, err := svc.listZiyanRuleHealthAndCert(cts.Kit, ruleIds)
	if err != nil {
		logs.Errorf("fail to list health and cert of tcloud url rule, err: %s, ruleIds: %v, rid: %s",
			err, ruleIds, cts.Kit.Rid)
		return nil, err
	}

	return svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		for _, rule := range req.UrlRules {
			update := &tablelb.TCloudZiyanLbUrlRuleTable{
				Name:               rule.Name,
				Domain:             rule.Domain,
				URL:                rule.URL,
				TargetGroupID:      rule.TargetGroupID,
				CloudTargetGroupID: rule.CloudTargetGroupID,
				Scheduler:          rule.Scheduler,
				Region:             rule.Region,
				SessionExpire:      converter.PtrToVal(rule.SessionExpire),
				SessionType:        rule.SessionType,
				Memo:               rule.Memo,
				Reviser:            cts.Kit.User,
			}

			if rule.HealthCheck != nil {
				hc := healthCertMap[rule.ID]
				mergedHealth, err := json.UpdateMerge(rule.HealthCheck, string(hc.Health))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge rule health check failed, err: %v", err)
				}
				update.HealthCheck = tabletype.JsonField(mergedHealth)

			}
			if rule.Certificate != nil {
				hc := healthCertMap[rule.ID]
				mergedCert, err := json.UpdateMerge(rule.Certificate, string(hc.Cert))
				if err != nil {
					return nil, fmt.Errorf("json UpdateMerge rule cert failed, err: %v", err)
				}
				update.Certificate = tabletype.JsonField(mergedCert)
			}
			err = svc.dao.LoadBalancerTCloudZiyanUrlRule().UpdateByIDWithTx(cts.Kit, txn, rule.ID, update)
			if err != nil {
				logs.Errorf("update tcloud rule by id failed, err: %v, id: %s, rid: %s", err, rule.ID, cts.Kit.Rid)
				return nil, fmt.Errorf("update rule failed, err: %v", err)
			}
		}

		return nil, nil
	})
}

func (svc *lbSvc) listZiyanRuleHealthAndCert(kt *kit.Kit, ruleIds []string) (map[string]tcloudHealthCert, error) {
	opt := &types.ListOption{
		Filter: tools.ContainersExpression("id", ruleIds),
		Page:   &core.BasePage{Limit: core.DefaultMaxPageLimit},
	}

	resp, err := svc.dao.LoadBalancerTCloudZiyanUrlRule().List(kt, opt)
	if err != nil {
		return nil, err
	}

	return converter.SliceToMap(resp.Details, func(t *tablelb.TCloudZiyanLbUrlRuleTable) (string, tcloudHealthCert) {
		return t.ID, tcloudHealthCert{Health: t.HealthCheck, Cert: t.Certificate}
	}), nil
}

func (svc *lbSvc) listTCloudZiyanLoadBalancerUrlRuleByTgIDs(kt *kit.Kit,
	lblReq protocloud.ListenerQueryItem, cloudClbIDs, cloudLblIDs, targetGroupIDs []string) (
	[]protocloud.LoadBalancerUrlRuleResult, error) {

	lblTargetFilter := make([]*filter.AtomRule, 0)
	lblTargetFilter = append(lblTargetFilter, tools.RuleIn("cloud_lb_id", cloudClbIDs))
	lblTargetFilter = append(lblTargetFilter, tools.RuleIn("cloud_lbl_id", cloudLblIDs))
	if len(targetGroupIDs) > 0 {
		lblTargetFilter = append(lblTargetFilter, tools.RuleIn("target_group_id", targetGroupIDs))
	}
	if len(lblReq.RuleType) > 0 {
		lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("rule_type", lblReq.RuleType))
		if lblReq.RuleType == enumor.Layer7RuleType {
			if len(lblReq.Domain) > 0 {
				lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("domain", lblReq.Domain))
			}
			if len(lblReq.Url) > 0 {
				lblTargetFilter = append(lblTargetFilter, tools.RuleEqual("url", lblReq.Url))
			}
		}
	}
	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(lblTargetFilter...),
		Page:   core.NewDefaultBasePage(),
	}
	lblTargetList := make([]protocloud.LoadBalancerUrlRuleResult, 0)
	for {
		loopLblTargetList, err := svc.dao.LoadBalancerTCloudZiyanUrlRule().List(kt, opt)
		if err != nil {
			logs.Errorf("list load balancer tcloud-ziyan url rule failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("list load balancer tcloud-ziyan url rule failed, err: %v", err)
		}

		for _, item := range loopLblTargetList.Details {
			urlRuleResult := protocloud.LoadBalancerUrlRuleResult{
				LbID:               item.LbID,
				CloudClbID:         item.CloudLbID,
				LblID:              item.LblID,
				CloudLblID:         item.CloudLBLID,
				TargetGroupRuleMap: make(map[string]protocloud.DomainUrlRuleInfo),
			}
			urlRuleResult.TargetGroupIDs = append(urlRuleResult.TargetGroupIDs, item.TargetGroupID)
			urlRuleResult.TargetGroupRuleMap[item.TargetGroupID] = protocloud.DomainUrlRuleInfo{
				RuleID:      item.ID,
				CloudRuleID: item.CloudID,
				RuleType:    item.RuleType,
				Domain:      item.Domain,
				Url:         item.URL,
			}
			lblTargetList = append(lblTargetList, urlRuleResult)
		}
		if uint(len(loopLblTargetList.Details)) < core.DefaultMaxPageLimit {
			break
		}

		opt.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
	return lblTargetList, nil
}
