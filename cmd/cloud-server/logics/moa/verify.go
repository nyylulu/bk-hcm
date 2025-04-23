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

package moa

import (
	"errors"
	"fmt"

	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	pkgmoa "hcm/pkg/thirdparty/moa"

	etcd3 "go.etcd.io/etcd/client/v3"
)

// Interface MOA Logic
type Interface interface {
	// RequestMoa 向MOA发起验证
	RequestMoa(kt *kit.Kit, scene enumor.MoaScene, affectedCount int, lang string) (sessionID string, err error)
	// VerifyMoa 向MOA查询验证结果
	VerifyMoa(kt *kit.Kit, scene enumor.MoaScene, sessionID string) (status enumor.MoaVerifyStatus, err error)
	// CheckCachedResult 检查本地是否有用户请求结果，如果已通过会删除结果
	CheckCachedResult(kt *kit.Kit, scene enumor.MoaScene, sessionID string) (err error)
}

// NewMoa ...
func NewMoa(moaCli pkgmoa.Client, etcdCli *etcd3.Client) Interface {
	return &moaLogic{
		moaCli:  moaCli,
		etcdCli: etcdCli,
	}
}

// GetMoaConfig ...
func GetMoaConfig(scene enumor.MoaScene) (cfgTpl cc.MoaTplCfg, isDefault bool) {
	cfgTpl = cc.CloudServer().MOA.DefaultTemplate
	isDefault = true
	for i := range cc.CloudServer().MOA.Templates {
		tpl := cc.CloudServer().MOA.Templates[i]
		if tpl.Scene != scene {
			continue
		}
		isDefault = false
		if tpl.Channel != "" {
			cfgTpl.Channel = tpl.Channel
		}
		if tpl.Timeout != 0 {
			cfgTpl.Timeout = tpl.Timeout
		}
		cfgTpl.ZH = tpl.ZH.Over(&cfgTpl.ZH)
		cfgTpl.EN = tpl.ZH.Over(&cfgTpl.EN)
		// only first match
		break
	}

	return cfgTpl, isDefault
}

// Logic Moa Logic
type moaLogic struct {
	moaCli  pkgmoa.Client
	etcdCli *etcd3.Client
}

// RequestMoa 发起Moa二次验证，会将对应SessionID存入etcd
func (m *moaLogic) RequestMoa(kt *kit.Kit, scene enumor.MoaScene, affectedCount int, lang string) (
	sessionID string, err error) {

	moaCfg, isDefault := GetMoaConfig(scene)
	if isDefault {
		return "", fmt.Errorf("invalid scene: %v", scene)
	}
	payload, err := moaCfg.BuildPayload(affectedCount)
	if err != nil {
		logs.Errorf("build payload failed, err: %v, scene: %v, count: %d, rid: %s", err, scene, affectedCount, kt.Rid)
		return "", fmt.Errorf("build payload failed, err: %v", err)
	}

	opt := &pkgmoa.InitiateVerificationReq{
		Username:      kt.User,
		Channel:       moaCfg.Channel,
		Language:      lang,
		PromptPayload: payload,
	}
	resp, err := m.moaCli.Request(kt, opt)
	if err != nil {
		logs.Errorf("request moa api failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return "", err
	}
	// 存入etcd
	err = m.saveRequest(kt, scene, resp.SessionId, int64(moaCfg.Timeout.Seconds()), enumor.MoaVerifyPending)
	if err != nil {
		return "", err
	}
	return resp.SessionId, nil
}

func (m *moaLogic) saveRequest(kt *kit.Kit, scene enumor.MoaScene, sessionID string, ttl int64,
	status enumor.MoaVerifyStatus) error {

	etcdKey := GetMoaKey(kt.User, scene, sessionID)
	leaseResp, err := m.etcdCli.Grant(kt.Ctx, ttl)
	if err != nil {
		logs.Errorf("grant etcd lease failed, err: %v, leaseResp: %v, rid: %s", err, leaseResp, kt.Rid)
		return err
	}
	putResp, err := m.etcdCli.Put(kt.Ctx, etcdKey, string(status), etcd3.WithLease(leaseResp.ID))
	if err != nil {
		logs.Errorf("put moa key with lease failed, err: %v, putResp: %v, rid: %s", err, putResp, kt.Rid)
		return err
	}
	return nil
}

func (m *moaLogic) getCachedStatus(kt *kit.Kit, scene enumor.MoaScene, sessionID string) (
	status enumor.MoaVerifyStatus, err error) {

	etcdKey := GetMoaKey(kt.User, scene, sessionID)

	getResp, err := m.etcdCli.Get(kt.Ctx, etcdKey)
	if err != nil {
		logs.Errorf("get moa key, err: %v, key: %s, rid: %s", err, etcdKey, kt.Rid)
		return "", err
	}
	if getResp.Count == 0 {
		return enumor.MoaVerifyNotFound, nil
	}

	return enumor.MoaVerifyStatus(getResp.Kvs[0].Value), nil
}

// GetMoaKey ...
func GetMoaKey(user string, scene enumor.MoaScene, sessionID string) string {
	return fmt.Sprintf("/moa/%s/%s/%s", user, scene, sessionID)
}

// VerifyMoa 查询moa校验结果，要求已有本地缓存的MOA结果，pending态会尝试向moa查询结果
func (m *moaLogic) VerifyMoa(kt *kit.Kit, scene enumor.MoaScene, sessionID string) (
	status enumor.MoaVerifyStatus, err error) {

	status, err = m.getCachedStatus(kt, scene, sessionID)
	if err != nil {
		return "", err
	}
	if status != enumor.MoaVerifyPending {
		return status, nil
	}
	opt := &pkgmoa.VerificationReq{
		SessionId: sessionID,
		Username:  kt.User,
	}
	resp, err := m.moaCli.Verify(kt, opt)
	if err != nil {
		return "", err
	}

	if resp.Status == enumor.MoaStatusPending {
		return enumor.MoaVerifyPending, nil
	}
	if resp.Status != enumor.MoaStatusFinish {
		logs.Errorf("unknown moa status: %s, button type: %s, scene: %s, rid: %s",
			resp.Status, resp.ButtonType, scene, kt.Rid)
		return "", fmt.Errorf("unknown moa status: %s", resp.Status)
	}

	resultStatus := enumor.MoaVerifyRejected
	if resp.ButtonType == enumor.MoaButtonTypeConfirm {
		resultStatus = enumor.MoaVerifyConfirmed
	} else if resp.ButtonType == enumor.MoaButtonTypeCancel {
		resultStatus = enumor.MoaVerifyRejected
	} else {
		logs.Errorf("unknown moa button type: %s, status: %s, rid: %s", resp.ButtonType, resp.Status, kt.Rid)
		return "", fmt.Errorf("unknown moa button type: %s", resp.ButtonType)
	}
	moaCfg, _ := GetMoaConfig(scene)
	err = m.saveRequest(kt, scene, sessionID, int64(moaCfg.Timeout.Seconds()), resultStatus)
	if err != nil {
		return "", err
	}
	return resultStatus, nil

}

// CheckCachedResult 检查本地缓存的MOA结果，如果已通过会删除结果(只能检查一次)
func (m *moaLogic) CheckCachedResult(kt *kit.Kit, scene enumor.MoaScene, sessionID string) error {

	status, err := m.getCachedStatus(kt, scene, sessionID)
	if err != nil {
		return err
	}

	if status != enumor.MoaVerifyConfirmed {
		switch status {

		case enumor.MoaVerifyPending:
			return errors.New("moa verify pending")
		case enumor.MoaVerifyRejected:
			return errors.New("moa verify rejected")
		case enumor.MoaVerifyNotFound:
			return errf.New(errf.MOAValidationTimeoutError, "moa session id expired or not found")
		default:
			return fmt.Errorf("unknown moa status: %s", status)
		}
	}

	// 对于已通过的情况，尝试删除缓存
	etcdKey := GetMoaKey(kt.User, scene, sessionID)

	rsp, err := m.etcdCli.Txn(kt.Ctx).
		If(etcd3.Compare(etcd3.Version(etcdKey), ">", 0)).
		Then(etcd3.OpDelete(etcdKey)).
		Commit()
	if err != nil {
		return err
	}
	if !rsp.Succeeded {
		// 抢失败，认为是已过期
		return errf.New(errf.MOAValidationTimeoutError, "moa session id expired or not found")
	}

	return nil
}
