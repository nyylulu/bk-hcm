/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package returner ...
package returner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	recovertask "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/erpapi"
)

func (r *Returner) returnPm(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) (string, error) {
	assetIds := make([]string, 0)
	for _, host := range hosts {
		if host.AssetID == "" {
			logs.Warnf("invalid host %s with empty asset id, rid: %s", host.AssetID, kt.Rid)
			continue
		}
		assetIds = append(assetIds, host.AssetID)
	}

	if len(assetIds) == 0 {
		logs.Errorf("failed to create device return order, for asset list is empty, rid: %s", kt.Rid)
		return "", fmt.Errorf("failed to create device return order, for asset list is empty")
	}

	retReason := erpapi.ReturnReasonRegular
	switch task.RecycleType {
	case table.RecycleTypeDissolve:
		retReason = erpapi.ReturnReasonDissolve
	case table.RecycleTypeExpired:
		// treat expired pm as ERP "常规回收"
		retReason = erpapi.ReturnReasonRegular
	}

	// default not emergent
	isEmergent := 0
	if task.ReturnPlan == table.RetPlanImmediate {
		isEmergent = 1
	}

	// default not skip double confirm
	skipConfirm := 0
	if task.SkipConfirm == true {
		skipConfirm = 1
	}

	// construct erp pm return request
	req := &erpapi.ErpReq{
		Params: &erpapi.ErpParam{
			Content: &erpapi.Content{
				Type:    erpapi.ReqType,
				Version: erpapi.ReqVersion,
				ReqInfo: &erpapi.ReqInfo{
					ReqKey:    erpapi.ReqKey,
					ReqModule: erpapi.ReqModule,
					// TODO: get from config
					Operator: erpapi.ReqOperator,
				},
				ReqItem: &erpapi.ReqItem{
					Method: erpapi.DeviceReturnMethod,
					Data: &erpapi.ReturnReqData{
						DeptId:      erpapi.IEGDeptId,
						AssetList:   assetIds,
						IsEmergent:  isEmergent,
						SkipConfirm: skipConfirm,
						Reason:      retReason,
						ReasonMsg:   "",
						Remark:      "",
					},
				},
			},
		},
	}

	resp, err := r.erp.CreateDeviceReturnOrder(nil, nil, req)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:returnPm:failed, failed to create device return order, err: %v, rid: %s",
			err, kt.Rid)
		return "", err
	}

	respStr := ""
	if b, err := json.Marshal(resp); err == nil {
		respStr = string(b)
	}

	logs.Infof("recycler:logics:cvm:returnPm:success, return pm resp: %s, rid: %s", respStr, kt.Rid)

	return r.parseReturnPmResp(kt, resp)
}

func (r *Returner) parseReturnPmResp(kt *kit.Kit, resp *erpapi.ErpResp) (string, error) {
	if resp.DataSet.Header.Code != 0 {
		logs.Errorf("pm return task failed, code: %d, msg: %s, rid: %s", resp.DataSet.Header.Code,
			resp.DataSet.Header.ErrMsg, kt.Rid)
		return "", fmt.Errorf("pm return task failed, code: %d, msg: %s", resp.DataSet.Header.Code,
			resp.DataSet.Header.ErrMsg)
	}

	bytes, err := json.Marshal(resp.DataSet.Data)
	if err != nil {
		logs.Errorf("pm return task failed, for parse pm return response err: %v, rid: %s", err, kt.Rid)
		return "", fmt.Errorf("pm return task failed, for parse pm return response err: %v", err)
	}

	var retData erpapi.ReturnRespData
	if err := json.Unmarshal(bytes, &retData); err != nil {
		logs.Errorf("pm return task failed, for parse pm return response err: %v, rid: %s", err, kt.Rid)
		return "", fmt.Errorf("pm return task failed, for parse pm return response err: %v", err)
	}

	if retData.OrderId == "" {
		logs.Errorf("pm return task failed, for order id is empty, rid: %s", kt.Rid)
		return "", fmt.Errorf("pm return task return empty order id")
	}

	return retData.OrderId, nil
}

func (r *Returner) queryPmOrder(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) *event.Event {
	// construct erp pm return request
	req := &erpapi.ErpReq{
		Params: &erpapi.ErpParam{
			Content: &erpapi.Content{
				Type:    erpapi.ReqType,
				Version: erpapi.ReqVersion,
				ReqInfo: &erpapi.ReqInfo{
					ReqKey:    erpapi.ReqKey,
					ReqModule: erpapi.ReqModule,
					// TODO: get from config
					Operator: erpapi.ReqOperator,
				},
				ReqItem: &erpapi.ReqItem{
					Method: erpapi.QueryReturnMethod,
					Data: &erpapi.OrderQueryReqData{
						OrderId: task.TaskID,
					},
				},
			},
		},
	}

	resp, err := r.erp.QueryDeviceReturnOrders(nil, nil, req)
	if err != nil {
		// keep loop query when error occurs until timeout
		logs.Warnf("failed to query pm return detail, err: %v, suborderID: %s, taskID: %s, rid: %s",
			err, task.SuborderID, task.TaskID, kt.Rid)
		return &event.Event{Type: event.ReturnHandling, Error: err}
	}

	respStr := ""
	if b, err := json.Marshal(resp); err == nil {
		respStr = string(b)
	}

	logs.Infof("query pm return detail suborderID: %s, taskID: %s, resp: %s, rid: %s",
		task.SuborderID, task.TaskID, respStr, kt.Rid)

	if resp.DataSet.Header.Code != 0 {
		// keep loop query when error occurs until timeout
		logs.Warnf("failed to query pm return detail, suborderID: %s, taskID: %s, code: %d, msg: %s, rid: %s",
			task.SuborderID, task.TaskID, resp.DataSet.Header.Code, resp.DataSet.Header.ErrMsg, kt.Rid)
		ev := &event.Event{
			Type: event.ReturnHandling,
			Error: fmt.Errorf("failed to query pm return detail, code: %d, msg: %s", resp.DataSet.Header.Code,
				resp.DataSet.Header.ErrMsg),
		}
		return ev
	}

	bytes, err := json.Marshal(resp.DataSet.Data)
	if err != nil {
		logs.Errorf("query pm return detail failed, for parse pm return response err: %v, rid: %s", err, kt.Rid)
		ev := &event.Event{
			Type:  event.ReturnFailed,
			Error: fmt.Errorf("query pm return detail failed, for parse pm return response err: %v", err),
		}
		return ev
	}

	var retData erpapi.OrderQueryRespData
	if err = json.Unmarshal(bytes, &retData); err != nil {
		logs.Errorf("query pm return detail failed, for parse pm return response err: %v, rid: %s", err, kt.Rid)
		ev := &event.Event{
			Type:  event.ReturnFailed,
			Error: fmt.Errorf("query pm return detail failed, for parse pm return response err: %v", err),
		}
		return ev
	}

	return r.processPmReturnResult(kt, task, hosts, retData)
}

func (r *Returner) processPmReturnResult(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost,
	retData erpapi.OrderQueryRespData) *event.Event {

	successCnt, failedCnt, runningCnt, isApproving, isRejected := r.parsePmReturnDetail(kt, hosts, retData.ResultSet)
	if runningCnt > 0 {
		handler := "AUTO"
		msg := ""
		if isApproving {
			handler = recovertask.Handler
			msg = "return order is approving"
		}
		if err := r.UpdateOrderInfo(context.Background(), task.SuborderID, handler, successCnt, failedCnt, runningCnt,
			msg); err != nil {
			logs.Warnf("failed to update recycle order %s info, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
			// ignore update error and continue to query
		}
		return &event.Event{Type: event.ReturnHandling, Error: nil}
	}

	if failedCnt > 0 {
		msg := fmt.Sprintf("%d hosts return failed", failedCnt)

		// transfer hosts back to recycle module if return order is rejected
		if isRejected {
			msg = "return order is rejected, hosts are transited back to recycle module"
			r.rollbackTransit(kt, hosts)
		}

		if err := r.UpdateReturnTaskInfo(context.Background(), task, "", table.ReturnStatusFailed, msg); err != nil {
			logs.Errorf("failed to update return task info, order id: %s, err: %v, rid: %s",
				task.SuborderID, err, kt.Rid)
			return &event.Event{Type: event.ReturnFailed, Error: err}
		}
		if err := r.UpdateOrderInfo(context.Background(), task.SuborderID, "AUTO", successCnt, failedCnt, runningCnt,
			msg); err != nil {
			logs.Warnf("failed to update recycle order %s info, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
			// ignore update error and continue to query
		}

		return &event.Event{Type: event.ReturnFailed, Error: nil}
	}

	if err := r.UpdateReturnTaskInfo(context.Background(), task, "", table.ReturnStatusSuccess, "success"); err != nil {
		logs.Errorf("failed to update return task info, order id: %s, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		return &event.Event{Type: event.ReturnFailed, Error: err}
	}

	if err := r.UpdateOrderInfo(context.Background(), task.SuborderID, "AUTO", successCnt, failedCnt, runningCnt,
		"success"); err != nil {
		logs.Warnf("failed to update recycle order %s info, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		// ignore update error and continue to query
	}

	return &event.Event{Type: event.ReturnSuccess, Error: nil}
}

func (r *Returner) parsePmReturnDetail(kt *kit.Kit, hosts []*table.RecycleHost, details []*erpapi.OrderQueryRst) (
	uint, uint, uint, bool, bool) {

	mapAsset2Detail := make(map[string]*erpapi.OrderQueryRst)
	for _, detail := range details {
		mapAsset2Detail[detail.AssetId] = detail
	}

	runningCnt := uint(0)
	failedCnt := uint(0)
	successCnt := uint(0)
	isApproving := false
	isRejected := false
	for _, host := range hosts {
		switch host.Status {
		case table.RecycleStatusDone:
			successCnt++
		case table.RecycleStatusReturnFailed:
			failedCnt++
		case table.RecycleStatusReturning:
			{
				detail, ok := mapAsset2Detail[host.AssetID]
				if !ok {
					runningCnt++
					continue
				}
				if detail.Status == 7 {
					isApproving = true
				} else if detail.Status == 9 {
					isRejected = true
				}
				if detail.RecycleStatus == 7 {
					// success
					successCnt++
					if err := r.updatePmHostInfo(kt, host, detail); err != nil {
						logs.Warnf("failed to update recycle host info, err: %v, rid: %s", err, kt.Rid)
					}
				} else if detail.RecycleStatus == 6 || detail.RecycleStatus == 8 ||
					detail.Status == 3 || detail.Status == 6 || detail.Status == 9 || detail.Status == 12 ||
					detail.CheckStatus == 3 {
					// failed
					failedCnt++
					if err := r.updatePmHostInfo(kt, host, detail); err != nil {
						logs.Warnf("failed to update recycle host info, err: %v, rid: %s", err, kt.Rid)
					}
				} else {
					// running
					runningCnt++
				}
			}
		default:
			logs.Warnf("%s query pm return detail failed, for invalid recycle host status %s, rid: %s",
				host.IP, host.Status, kt.Rid)
			failedCnt++
		}
	}

	return successCnt, failedCnt, runningCnt, isApproving, isRejected
}

func (r *Returner) updatePmHostInfo(kt *kit.Kit, host *table.RecycleHost, detail *erpapi.OrderQueryRst) error {
	filter := mapstr.MapStr{
		"suborder_id": host.SuborderID,
		"ip":          host.IP,
	}

	now := time.Now()
	update := mapstr.MapStr{
		"return_tag":       detail.OBSLabel,
		"return_cost_rate": 0.0,
		"update_at":        now,
	}

	if detail.RecycleStatus == 7 {
		update["stage"] = table.RecycleStageDone
		update["status"] = table.RecycleStatusDone
		update["return_time"] = now.Format("2006-01-02 15:04:05")
	} else if detail.RecycleStatus == 6 || detail.RecycleStatus == 8 {
		update["status"] = table.RecycleStatusReturnFailed
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(context.Background(), &filter, &update); err != nil {
		logs.Errorf("failed to update recycle host, ip: %s, err: %v, rid: %s", host.IP, err, kt.Rid)
		return err
	}

	return nil
}
