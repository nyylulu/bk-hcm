/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package detector ...
package detector

import (
	"fmt"
	"strings"
	"time"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/thirdparty/tmpapi"
	"hcm/pkg/logs"

	"github.com/mitchellh/mapstructure"
)

func (d *Detector) basicCheck(step *table.DetectStep, retry int) (int, string, error) {
	attempt := 0
	exeInfo := ""
	var err error = nil

	for i := 0; i < retry; i++ {
		attempt = i
		exeInfo, err = d.checkStrategies(step.IP)
		if err == nil {
			break
		}

		// retry gap until last retry
		if (i + 1) < retry {
			time.Sleep(3 * time.Second)
		}
	}

	return attempt, exeInfo, err
}

func (d *Detector) checkStrategies(ip string) (string, error) {
	exeInfos := make([]string, 0)

	// check tmp
	tmpExeInfo, errTmp := d.checkTmp(ip)
	exeInfos = append(exeInfos, tmpExeInfo)
	if errTmp != nil {
		return strings.Join(exeInfos, "\n"), errTmp
	}

	// check tgw
	tgwExeInfo, errTgw := d.checkTgw(ip)
	exeInfos = append(exeInfos, tgwExeInfo)
	if errTgw != nil {
		return strings.Join(exeInfos, "\n"), errTgw
	}

	// check tgw nat
	tgwNatExeInfo, errTgwNat := d.checkTgwNat(ip)
	exeInfos = append(exeInfos, tgwNatExeInfo)
	if errTgwNat != nil {
		return strings.Join(exeInfos, "\n"), errTgwNat
	}

	// check L5
	l5ExeInfo, errL5 := d.checkL5(ip)
	exeInfos = append(exeInfos, l5ExeInfo)
	if errL5 != nil {
		return strings.Join(exeInfos, "\n"), errL5
	}

	return strings.Join(exeInfos, "\n"), nil
}

func (d *Detector) checkTmp(ip string) (string, error) {
	exeInfos := make([]string, 0)

	respTmp, err := d.tmp.CheckTMP(nil, nil, ip)
	if err != nil {
		logs.Errorf("failed to check gcs, ip: %s, err: %v", ip, err)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check tmp, err: %v", err)
	}

	tmpRespStr := d.structToStr(respTmp)
	exeInfo := fmt.Sprintf("tmp response: %s", tmpRespStr)
	exeInfos = append(exeInfos, exeInfo)

	for _, aRule := range respTmp {
		rule := new(tmpapi.AlarmShieldConfig)
		if err := mapstructure.Decode(aRule, rule); err != nil {
			return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check tmp, err: %v", err)
		}

		if d.isAllShieldRule(rule) && d.isShieldByIp(rule) {
			if !d.isShieldEndTimeMoreThanThreeDay(rule) {
				continue
			}
			logs.Infof("%s has tmp shield strategy", ip)
			return strings.Join(exeInfos, "\n"), fmt.Errorf("has tmp shield policy")
		}
	}

	return strings.Join(exeInfos, "\n"), nil
}

func (d *Detector) checkTgw(ip string) (string, error) {
	exeInfos := make([]string, 0)

	respTgw, err := d.tgw.CheckTgw(nil, nil, ip)
	if err != nil {
		logs.Errorf("failed to check tgw, ip: %s, err: %v", ip, err)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check tgw, err: %v", err)
	}

	tgwRespStr := d.structToStr(respTgw)
	exeInfo := fmt.Sprintf("tgw response: %s", tgwRespStr)
	exeInfos = append(exeInfos, exeInfo)

	if respTgw.Errno != 0 {
		logs.Errorf("%s failed to check tgw, errno: %d, err: %s", ip, respTgw.Errno, respTgw.Error)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check tgw, errno: %d, err: %s", respTgw.Errno,
			respTgw.Error)
	}

	if respTgw.L7RuleCount > 0 || len(respTgw.L7RuleList) > 0 {
		logs.Infof("%s has %d stgw rules", ip, respTgw.L7RuleCount)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has %d stgw rules", respTgw.L7RuleCount)
	}

	if respTgw.RuleCount > 0 || len(respTgw.RuleList) > 0 {
		logs.Infof("%s has %d tgw rules", ip, respTgw.RuleCount)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has %d tgw rules", respTgw.RuleCount)
	}

	return strings.Join(exeInfos, "\n"), nil
}

func (d *Detector) checkTgwNat(ip string) (string, error) {
	exeInfos := make([]string, 0)

	respTgwNat, err := d.tgw.CheckTgwNat(nil, nil, ip)
	if err != nil {
		logs.Errorf("failed to check tgw nat, ip: %s, err: %v", ip, err)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check tgw nat, err: %v", err)
	}

	tgwNatRespStr := d.structToStr(respTgwNat)
	exeInfo := fmt.Sprintf("tgw nat response: %s", tgwNatRespStr)
	exeInfos = append(exeInfos, exeInfo)

	if respTgwNat.Errno != 0 {
		logs.Errorf("%s failed to check tgw nat, errno: %d, err: %s", ip, respTgwNat.Errno, respTgwNat.Error)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check tgw nat, errno: %d, err: %s", respTgwNat.Errno,
			respTgwNat.Error)
	}

	if respTgwNat.TotalCount > 0 || len(respTgwNat.RSList) > 0 {
		logs.Infof("%s has tgw nat policy", ip)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has tgw nat policy")
	}

	return strings.Join(exeInfos, "\n"), nil
}

func (d *Detector) checkL5(ip string) (string, error) {
	exeInfos := make([]string, 0)

	respL5, err := d.l5.CheckL5(nil, nil, ip)
	if err != nil {
		logs.Errorf("failed to check l5, ip: %s, err: %v", ip, err)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check l5, err: %v", err)
	}

	l5RespStr := d.structToStr(respL5)
	exeInfo := fmt.Sprintf("l5 response: %s", l5RespStr)
	exeInfos = append(exeInfos, exeInfo)

	if respL5.ReturnCode != 0 {
		logs.Errorf("%s failed to check l5, code: %d, msg: %s", ip, respL5.ReturnCode, respL5.ReturnMessage)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("failed to check l5, code: %d, msg: %s", respL5.ReturnCode,
			respL5.ReturnMessage)
	}

	if len(respL5.Data.SIDList) > 0 {
		logs.Infof("%s has l5 policy", ip)
		return strings.Join(exeInfos, "\n"), fmt.Errorf("has l5 policy")
	}

	return strings.Join(exeInfos, "\n"), nil
}

func (d *Detector) isAllShieldRule(rule *tmpapi.AlarmShieldConfig) bool {
	return rule.ShieldRule == `["true"]`
}

func (d *Detector) isShieldByIp(rule *tmpapi.AlarmShieldConfig) bool {
	return strings.HasPrefix(rule.CiSetInfo, "<br>IP:")
}

func (d *Detector) isShieldEndTimeMoreThanThreeDay(rule *tmpapi.AlarmShieldConfig) bool {
	end, _ := time.Parse("2006-01-02", rule.CycleEnd)
	add, _ := time.ParseDuration("72h")
	if time.Now().Add(add).After(end) {
		return false
	}

	return true
}
