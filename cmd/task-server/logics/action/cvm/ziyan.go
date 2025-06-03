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

package actioncvm

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	actcli "hcm/cmd/task-server/logics/action/cli"
	actionflow "hcm/cmd/task-server/logics/flow"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	coretask "hcm/pkg/api/core/task"
	protocloud "hcm/pkg/api/data-service/cloud"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	cvmproto "hcm/pkg/api/task-server/cvm"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/alarmapi"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/sampwdapi"
	"hcm/pkg/tools/retry"
)

func (c StartActionV2) startTCloudZiyanCvm(kt *kit.Kit, opt *cvmproto.CvmOperationOption) error {
	req := &hcprotocvm.TCloudBatchStartReq{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		IDs:       opt.IDs,
	}
	executeErr := actcli.GetHCService().TCloudZiyan.Cvm.BatchStartCvm(kt, req)
	if executeErr != nil {
		logs.Errorf("fail to call hc to start cvms, err: %v, req: %+v, rid: %s",
			executeErr, opt, kt.Rid)
		err := actionflow.BatchUpdateTaskDetailResultState(
			kt, opt.ManagementDetailIDs, enumor.TaskDetailFailed, nil, executeErr)
		if err != nil {
			logs.Errorf("fail to set detail to failed after cloud operation, err: %v, rid: %s",
				err, kt.Rid)
		}
		return err
	}

	// 更新任务状态为 success
	err := actionflow.BatchUpdateTaskDetailResultState(kt, opt.ManagementDetailIDs, enumor.TaskDetailSuccess, nil, nil)
	if err != nil {
		logs.Errorf("fail to set detail to success after cloud operation, err: %v, rid: %s",
			err, kt.Rid)
		return err
	}
	return nil
}

func (c StopActionV2) stopTCloudZiyanCvm(kt *kit.Kit, opt *cvmproto.CvmOperationOption) error {
	req := &hcprotocvm.TCloudBatchStopReq{
		AccountID:   opt.AccountID,
		Region:      opt.Region,
		IDs:         opt.IDs,
		StopType:    typecvm.SoftFirst,
		StoppedMode: typecvm.KeepCharging,
	}
	executeErr := actcli.GetHCService().TCloudZiyan.Cvm.BatchStopCvm(kt, req)
	if executeErr != nil {
		logs.Errorf("fail to call hc to start cvms, err: %v, req: %+v, rid: %s",
			executeErr, opt, kt.Rid)
		err := actionflow.BatchUpdateTaskDetailResultState(
			kt, opt.ManagementDetailIDs, enumor.TaskDetailFailed, nil, executeErr)
		if err != nil {
			logs.Errorf("fail to set detail to failed after cloud operation, err: %v, rid: %s",
				err, kt.Rid)
		}
		return err
	}

	// 更新任务状态为 success
	err := actionflow.BatchUpdateTaskDetailResultState(kt, opt.ManagementDetailIDs, enumor.TaskDetailSuccess, nil, nil)
	if err != nil {
		logs.Errorf("fail to set detail to success after cloud operation, err: %v, rid: %s",
			err, kt.Rid)
		return err
	}
	return nil
}

func (c RebootActionV2) rebootTCloudZiyanCvm(kt *kit.Kit, opt *cvmproto.CvmOperationOption) error {
	req := &hcprotocvm.TCloudBatchRebootReq{
		AccountID: opt.AccountID,
		Region:    opt.Region,
		IDs:       opt.IDs,
		StopType:  typecvm.SoftFirst,
	}
	executeErr := actcli.GetHCService().TCloudZiyan.Cvm.BatchRebootCvm(kt, req)
	if executeErr != nil {
		logs.Errorf("fail to call hc to start cvms, err: %v, req: %+v, rid: %s",
			executeErr, opt, kt.Rid)
		err := actionflow.BatchUpdateTaskDetailResultState(
			kt, opt.ManagementDetailIDs, enumor.TaskDetailFailed, nil, executeErr)
		if err != nil {
			logs.Errorf("fail to set detail to failed after cloud operation, err: %v, rid: %s",
				err, kt.Rid)
		}
		return err
	}

	// 更新任务状态为 success
	err := actionflow.BatchUpdateTaskDetailResultState(kt, opt.ManagementDetailIDs, enumor.TaskDetailSuccess, nil, nil)
	if err != nil {
		logs.Errorf("fail to set detail to success after cloud operation, err: %v, rid: %s",
			err, kt.Rid)
		return err
	}
	return nil
}

func (act BatchTaskCvmResetAction) resetTCloudZiyanCvm(kt *kit.Kit, detail coretask.Detail,
	req *hcprotocvm.TCloudBatchResetCvmReq) error {

	cvms, err := getZiyanCvmInfo(kt, req.CloudIDs)
	if err != nil {
		return err
	}

	if err = validateCvmSvrStatus(kt, cvms, detail); err != nil {
		logs.Errorf("fail to validate cvm status, err: %v, req: %+v, rid: %s")
		return err
	}

	// 屏蔽告警
	serverBindIP := cc.TaskServer().Network.BindIP
	alarmIDs, err := actcli.GetAlarmCli().AddShieldAlarm(kt, req.IPs, alarmapi.ShieldHour, serverBindIP, "")
	if err != nil {
		logs.Errorf("failed to add shield alarm, err: %v, ips: %v, rid: %s", err, req.IPs, kt.Rid)
		return err
	}

	// update cmdb cvm srv_status
	for _, cvm := range cvms {
		if err = updateCMDBCvmOSAndSvrStatus(kt, cvm.Extension.BkAssetID, constant.ResetingSrvStatus, ""); err != nil {
			logs.Errorf("update cmdb cvm os failed, err: %v, bkAssetID: %s, rid: %s",
				err, cvm.Extension.BkAssetID, kt.Rid)
			return err
		}
	}

	var cloudErr error
	rangeMS := [2]uint{constant.CvmBatchTaskRetryDelayMinMS, constant.CvmBatchTaskRetryDelayMaxMS}
	policy := retry.NewRetryPolicy(0, rangeMS)
	for {
		cloudErr = actcli.GetHCService().TCloudZiyan.Cvm.ResetCvm(kt, req)
		cvmResetJson, jsonErr := json.Marshal(req)
		if jsonErr != nil {
			logs.Errorf("call hcservice api reset cvm json marshal, vendor: %s, detailID: %s, taskManageID: %s, "+
				"flowID: %s, cvmResetJson: %s, err: %+v, jsonErr: %+v, rid: %s", req.Vendor, detail.ID,
				detail.TaskManagementID, detail.FlowID, cvmResetJson, err, jsonErr, kt.Rid)
			return jsonErr
		}
		// 仅在碰到限频错误时进行重试
		if cloudErr != nil && strings.Contains(cloudErr.Error(), constant.TCloudLimitExceededErrCode) {
			if policy.RetryCount()+1 < actionflow.BatchTaskDefaultRetryTimes {
				// 	非最后一次重试，继续sleep
				logs.Errorf("call tcloud-ziyan cvm reset reach rate limit, will sleep for retry, retry count: %d, "+
					"err: %v, rid: %s", policy.RetryCount(), cloudErr, kt.Rid)
				policy.Sleep()
				continue
			}
		}
		// 其他情况都跳过
		break
	}
	// 云端报错后，需要恢复主机状态、解除屏蔽，支持用户再次提交，所以放在下面再报错

	newImageName := ""
	if cloudErr == nil {
		newImageName = req.ImageName
	}
	for _, cvm := range cvms {
		// update cmdb cvm os info
		if err = updateCMDBCvmOSAndSvrStatus(kt, cvm.Extension.BkAssetID, cvm.Extension.SrvStatus,
			newImageName); err != nil {

			logs.Errorf("update cmdb cvm os failed, err: %v, bkAssetID: %s, cvmCloudID: %s, taskManageID: %s, "+
				"flowID: %s, rid: %s", err, cvm.CloudID, detail.TaskManagementID, detail.FlowID,
				cvm.Extension.BkAssetID, kt.Rid)
			return err
		}
	}

	// 解除屏蔽
	if len(alarmIDs) > 0 {
		_, err = actcli.GetAlarmCli().DelShieldAlarm(kt, alarmIDs, serverBindIP)
		if err != nil {
			// 解除屏蔽可忽略，不影响主功能
			logs.Warnf("failed to del shield alarm, err: %v, alarmIDs: %v, ips: %v, rid: %s",
				err, alarmIDs, req.IPs, kt.Rid)
		}
	}

	// 记录云端报错信息
	if cloudErr != nil {
		logs.Errorf("failed to call hcservice to reset cvm, err: %v, detailID: %s, taskManageID: %s, flowID: %s, "+
			"rid: %s", cloudErr, detail.ID, detail.TaskManagementID, detail.FlowID, kt.Rid)
		return cloudErr
	}

	// 更新主机密码
	errMap := act.updateHostPwd(kt, cvms, detail.Creator, req.Password)
	if len(errMap) > 0 {
		// 更新失败不影响主流程，记录告警日志
		logs.Errorf("%s: failed to update host pwd, detailID: %s, taskManageID: %s, flowID: %s, errMap: %+v, rid: %s",
			constant.CvmResetSystemUpdatePwdFailed, detail.ID, detail.TaskManagementID, detail.FlowID, errMap, kt.Rid)
		return fmt.Errorf("主机重装成功，但iegtjj密码库更新失败(一般是由于主机的主备负责人发生变化)")
	}

	return nil
}

func validateCvmSvrStatus(kt *kit.Kit, cvms []corecvm.Cvm[corecvm.TCloudZiyanHostExtension],
	detail coretask.Detail) error {

	// get cvm info from cc, and check the srv_status isn't resetting
	mapBizToHostIDs := make(map[int64][]int64)
	for _, cvm := range cvms {
		mapBizToHostIDs[cvm.BkBizID] = append(mapBizToHostIDs[cvm.BkBizID], cvm.BkHostID)
	}
	for bizID, hostIDs := range mapBizToHostIDs {
		params := &cmdb.ListBizHostParams{
			BizID:  bizID,
			Fields: []string{"bk_host_id", "bk_host_innerip", "srv_status", "operator", "bk_bak_operator"},
			Page:   &cmdb.BasePage{Start: 0, Limit: int64(core.DefaultMaxPageLimit), Sort: "bk_host_id"},
			HostPropertyFilter: &cmdb.QueryFilter{
				Rule: &cmdb.CombinedRule{
					Condition: "AND",
					Rules: []cmdb.Rule{
						&cmdb.AtomRule{Field: "bk_host_id", Operator: "in", Value: hostIDs},
					},
				},
			},
		}
		hostResult, err := actcli.GetCMDBCli().ListBizHost(kt, params)
		if err != nil {
			logs.Errorf("request cmdb list biz host failed, err: %v, bizID: %d, hostIDs: %v, rid: %s",
				err, bizID, hostIDs, kt.Rid)
			return err
		}

		for _, host := range hostResult.Info {
			logs.Infof("cvm reset check status loop, hostID: %d, srv_status: %s, hostInnerIP: %s, operator: %s, "+
				"bkOperator: %s, detailCreator: %s, rid: %s", host.BkHostID, host.SrvStatus, host.BkHostInnerIP,
				host.Operator, host.BkBakOperator, detail.Creator, kt.Rid)
			// 校验主备负责人
			if !strings.Contains(host.Operator, detail.Creator) &&
				!strings.Contains(host.BkBakOperator, detail.Creator) {

				logs.Errorf("cvm reset check operator failed, hostID: %d, 重装的主机负责人校验失败："+
					"主机[%s]的主要负责人[%s]、备份负责人[%s]和当前任务执行人[%s]不匹配，请重新校验后提交, rid: %s",
					host.BkHostID, host.BkHostInnerIP, host.Operator, host.BkBakOperator, detail.Creator, kt.Rid)
				return fmt.Errorf("重装的主机负责人校验失败：主机[%s]的负责人和当前任务执行人[%s]不匹配，请重新校验后提交",
					host.BkHostInnerIP, detail.Creator)
			}
			if host.SrvStatus == constant.ResetingSrvStatus {
				logs.Errorf("cvm reset check srv status failed, hostID: %d, cvm is resetting, hostInnerIP: %s, rid: %s",
					host.BkHostID, host.BkHostInnerIP, kt.Rid)
				return fmt.Errorf("cvm srv_status is resetting, hostID: %d, hostInnerIP: %s, rid: %s",
					host.BkHostID, host.BkHostInnerIP, kt.Rid)
			}
		}
	}
	return nil
}

func getZiyanCvmInfo(kt *kit.Kit, cloudIDs []string) ([]corecvm.Cvm[corecvm.TCloudZiyanHostExtension], error) {
	listReq := &protocloud.CvmListReq{
		Filter: tools.ContainersExpression("cloud_id", cloudIDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := actcli.GetDataService().TCloudZiyan.Cvm.ListCvmExt(kt.Ctx, kt.Header(), listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud-ziyan cvm failed, err: %v, cloudID: %s, rid: %s",
			err, cloudIDs, kt.Rid)
		return nil, err
	}
	return listResp.Details, nil
}

func updateCMDBCvmOSAndSvrStatus(kt *kit.Kit, bkAssetID, srvStatus, osName string) error {
	updateReq := &cmdb.UpdateCvmOSReq{
		BkAssetId: bkAssetID,
		Data: cmdb.UpdateCvmOSReqData{
			SrvStatus: srvStatus,
		},
	}
	if osName != "" {
		updateReq.Data.BkOsName = osName
		updateReq.Data.BkOsVersion = "-"
	}

	err := actcli.GetCMDBCli().UpdateCvmOSAndSvrStatus(kt, updateReq)
	if err != nil {
		logs.Errorf("update cmdb cvm os failed, err: %v, bkAssetID: %s, rid: %s", err, bkAssetID, kt.Rid)
		return err
	}
	return nil
}

func (act BatchTaskCvmResetAction) updateHostPwd(kt *kit.Kit, cvms []corecvm.Cvm[corecvm.TCloudZiyanHostExtension],
	operator, pwd string) map[int]error {

	var errMap = make(map[int]error)
	for idx, cvm := range cvms {
		pwdReq := sampwdapi.UpdateHostPwdReq{
			BkHostID:     cvm.BkHostID,
			Password:     pwd,
			UserName:     operator, //  操作人
			GenerateTime: time.Now().Format(time.RFC3339),
		}
		rangeMS := [2]uint{constant.CvmBatchTaskRetryDelayMinMS, constant.CvmBatchTaskRetryDelayMaxMS}
		policy := retry.NewRetryPolicy(0, rangeMS)
		for {
			pwdResp, err := actcli.GetSamPwdCli().UpdateHostPwd(kt, pwdReq)
			if err != nil {
				if policy.RetryCount()+1 < actionflow.BatchTaskDefaultRetryTimes {
					// 	非最后一次重试，继续sleep
					logs.Errorf("failed to update host password, will sleep for retry, retry count: %d, "+
						"cvmCloudID: %s, hostID: %d, err: %+v, rid: %s",
						policy.RetryCount(), cvm.CloudID, cvm.BkHostID, err, kt.Rid)
					policy.Sleep()
					continue
				}

				errMap[idx] = fmt.Errorf("failed to update host password, cvmCloudID: %s, hostID: %d, IPs: %v, "+
					"err: %v", cvm.CloudID, cvm.BkHostID, cvm.PrivateIPv4Addresses, err)
				break
			}

			if pwdResp.ID <= 0 {
				logs.Errorf("failed to update host password id, cvmCloudID: %s, hostID: %d, pwdResp: %+v, rid: %s",
					cvm.CloudID, cvm.BkHostID, pwdResp, kt.Rid)
				errMap[idx] = fmt.Errorf("failed to update host password id, cvmCloudID: %s, hostID: %d, "+
					"IPs: %v, pwdResp: %+v", cvm.CloudID, cvm.BkHostID, cvm.PrivateIPv4Addresses, pwdResp)
			}
			break
		}
	}
	return errMap
}
