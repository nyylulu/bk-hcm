/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

// Package detector ...
package detector

import (
	"fmt"
	"strings"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/tmpapi"

	"github.com/mitchellh/mapstructure"
)

// checkBasicWorkGroup ...
type checkBasicWorkGroup struct {
	baseWorkGroup
}

// newCheckBasicWorkGroup ...
func newCheckBasicWorkGroup(resultHandler StepResultHandler, workerNum int, cliSet *cliSet) *checkBasicWorkGroup {
	return &checkBasicWorkGroup{
		baseWorkGroup: newBaseWorkGroup(enumor.CheckBasicDetectStep, resultHandler, workerNum, checkBasic, cliSet),
	}
}

// MaxBatchSize 最大批量数
func (c *checkBasicWorkGroup) MaxBatchSize() int {
	return onlyOneBatchSize
}

// checkBasic ...
func checkBasic(kt *kit.Kit, steps []*StepMeta, resultHandler StepResultHandler, cliSet *cliSet) {
	if len(steps) == 0 {
		logs.Warnf("check base worker receive empty steps, rid: %s", kt.Rid)
		return
	}

	for _, step := range steps {
		exeInfo, retry, err := checkStrategies(kt, step.Step.IP, cliSet)
		if err != nil {
			logs.Errorf("recycle detector check basic failed, err: %v, ip: %s, rid: %s", err, step.Step.IP, kt.Rid)
		}
		resultHandler.HandleResult(kt, []*StepMeta{step}, err, exeInfo, retry)
	}
}

func checkStrategies(kt *kit.Kit, ip string, cliSet *cliSet) (string, bool, error) {
	exeInfos := make([]string, 0)

	// check tmp
	tmpExeInfo, retry, errTmp := checkTmp(kt, ip, cliSet)
	exeInfos = append(exeInfos, tmpExeInfo)
	if errTmp != nil {
		return strings.Join(exeInfos, "\n"), retry, errTmp
	}

	// check tgw
	tgwExeInfo, retry, errTgw := checkTgw(kt, ip, cliSet)
	exeInfos = append(exeInfos, tgwExeInfo)
	if errTgw != nil {
		return strings.Join(exeInfos, "\n"), retry, errTgw
	}

	// check tgw nat
	tgwNatExeInfo, retry, errTgwNat := checkTgwNat(kt, ip, cliSet)
	exeInfos = append(exeInfos, tgwNatExeInfo)
	if errTgwNat != nil {
		return strings.Join(exeInfos, "\n"), retry, errTgwNat
	}

	// check L5
	l5ExeInfo, retry, errL5 := checkL5(kt, ip, cliSet)
	exeInfos = append(exeInfos, l5ExeInfo)
	if errL5 != nil {
		return strings.Join(exeInfos, "\n"), retry, errL5
	}

	return strings.Join(exeInfos, "\n"), false, nil
}

func checkTmp(kt *kit.Kit, ip string, cliSet *cliSet) (string, bool, error) {
	exeInfos := make([]string, 0)

	respTmp, err := cliSet.tmp.CheckTMP(kt.Ctx, kt.Header(), ip)
	if err != nil {
		logs.Errorf("recycle detector check basic tmp failed, err: %v, ip: %s, rid: %s", err, ip, kt.Rid)
		return strings.Join(exeInfos, "\n"), true, fmt.Errorf("failed to check tmp, err: %v", err)
	}

	tmpRespStr := structToStr(respTmp)
	exeInfo := fmt.Sprintf("tmp response: %s", tmpRespStr)
	exeInfos = append(exeInfos, exeInfo)

	for _, one := range respTmp {
		rule := new(tmpapi.AlarmShieldConfig)
		if err := mapstructure.Decode(one, rule); err != nil {
			logs.Errorf("recycle detector check basic tmp failed, failed to decode, err: %v, rid: %s", err, kt.Rid)
			return strings.Join(exeInfos, "\n"), true, fmt.Errorf("failed to check tmp, err: %v", err)
		}

		if isAllShieldRule(rule) && isShieldByIp(rule) {
			if !isShieldEndTimeMoreThanThreeDay(rule) {
				continue
			}
			logs.Infof("has tmp shield strategy, ip: %s, rid: %s", ip, kt.Rid)
			return strings.Join(exeInfos, "\n"), false, fmt.Errorf("has tmp shield policy")
		}
	}

	return strings.Join(exeInfos, "\n"), false, nil
}

func checkTgw(kt *kit.Kit, ip string, cliSet *cliSet) (string, bool, error) {
	exeInfos := make([]string, 0)

	respTgw, err := cliSet.tgw.CheckTgw(kt.Ctx, kt.Header(), ip)
	if err != nil {
		logs.Errorf("failed to check tgw, err: %v, ip: %s, rid: %s", err, ip, kt.Rid)
		return strings.Join(exeInfos, "\n"), true, fmt.Errorf("failed to check tgw, err: %v", err)
	}

	tgwRespStr := structToStr(respTgw)
	exeInfo := fmt.Sprintf("tgw response: %s", tgwRespStr)
	exeInfos = append(exeInfos, exeInfo)

	if respTgw.Errno != 0 {
		logs.Errorf("failed to check tgw, ip: %s, errno: %d, err: %s, rid: %s", ip, respTgw.Errno, respTgw.Error,
			kt.Rid)
		return strings.Join(exeInfos, "\n"), true, fmt.Errorf("failed to check tgw, errno: %d, err: %s", respTgw.Errno,
			respTgw.Error)
	}

	if respTgw.L7RuleCount > 0 || len(respTgw.L7RuleList) > 0 {
		logs.Infof("has stgw rules, ip: %s, count: %d, rid: %s", ip, respTgw.L7RuleCount, kt.Rid)
		return strings.Join(exeInfos, "\n"), false, fmt.Errorf("has %d stgw rules", respTgw.L7RuleCount)
	}

	if respTgw.RuleCount > 0 || len(respTgw.RuleList) > 0 {
		logs.Infof("has tgw rules, ip: %s, count: %d, rid: %s", ip, respTgw.RuleCount, kt.Rid)
		return strings.Join(exeInfos, "\n"), false, fmt.Errorf("has tgw rules, count: %d", respTgw.RuleCount)
	}

	return strings.Join(exeInfos, "\n"), false, nil
}

func checkTgwNat(kt *kit.Kit, ip string, cliSet *cliSet) (string, bool, error) {
	exeInfos := make([]string, 0)

	respTgwNat, err := cliSet.tgw.CheckTgwNat(kt.Ctx, kt.Header(), ip)
	if err != nil {
		logs.Errorf("failed to check tgw nat, ip: %s, err: %v, rid: %s", ip, err, kt.Rid)
		return strings.Join(exeInfos, "\n"), true, fmt.Errorf("failed to check tgw nat, err: %v", err)
	}

	tgwNatRespStr := structToStr(respTgwNat)
	exeInfo := fmt.Sprintf("tgw nat response: %s", tgwNatRespStr)
	exeInfos = append(exeInfos, exeInfo)

	if respTgwNat.Errno != 0 {
		logs.Errorf("failed to check tgw nat, ip: %s, errno: %d, err: %s, rid: %s", ip, respTgwNat.Errno,
			respTgwNat.Error, kt.Rid)
		return strings.Join(exeInfos, "\n"), true, fmt.Errorf("failed to check tgw nat, errno: %d, err: %s",
			respTgwNat.Errno, respTgwNat.Error)
	}

	if respTgwNat.TotalCount > 0 || len(respTgwNat.RSList) > 0 {
		logs.Infof("has tgw nat policy, ip: %s, rid: %s", ip, kt.Rid)
		return strings.Join(exeInfos, "\n"), false, fmt.Errorf("has tgw nat policy")
	}

	return strings.Join(exeInfos, "\n"), false, nil
}

func checkL5(kt *kit.Kit, ip string, cliSet *cliSet) (string, bool, error) {
	exeInfos := make([]string, 0)

	respL5, err := cliSet.l5.CheckL5(kt.Ctx, kt.Header(), ip)
	if err != nil {
		logs.Errorf("failed to check l5, ip: %s, err: %v, rid: %s", ip, err, kt.Rid)
		return strings.Join(exeInfos, "\n"), true, fmt.Errorf("failed to check l5, err: %v", err)
	}

	l5RespStr := structToStr(respL5)
	exeInfo := fmt.Sprintf("l5 response: %s", l5RespStr)
	exeInfos = append(exeInfos, exeInfo)

	if respL5.ReturnCode != 0 {
		logs.Errorf("failed to check l5, ip: %s, code: %d, msg: %s, rid: %s", ip, respL5.ReturnCode,
			respL5.ReturnMessage, kt.Rid)
		return strings.Join(exeInfos, "\n"), true, fmt.Errorf("failed to check l5, code: %d, msg: %s",
			respL5.ReturnCode, respL5.ReturnMessage)
	}

	if len(respL5.Data.SIDList) > 0 {
		logs.Infof("has l5 policy, ip: %s, rid: %s", ip, kt.Rid)
		return strings.Join(exeInfos, "\n"), false, fmt.Errorf("has l5 policy")
	}

	return strings.Join(exeInfos, "\n"), false, nil
}

func isAllShieldRule(rule *tmpapi.AlarmShieldConfig) bool {
	return rule.ShieldRule == `["true"]`
}

func isShieldByIp(rule *tmpapi.AlarmShieldConfig) bool {
	return strings.HasPrefix(rule.CiSetInfo, "<br>IP:")
}

func isShieldEndTimeMoreThanThreeDay(rule *tmpapi.AlarmShieldConfig) bool {
	end, _ := time.Parse("2006-01-02", rule.CycleEnd)
	add, _ := time.ParseDuration("72h")
	if time.Now().Add(add).After(end) {
		return false
	}

	return true
}
