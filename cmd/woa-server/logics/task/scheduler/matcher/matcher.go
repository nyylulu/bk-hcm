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

// Package matcher provides ...
package matcher

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"hcm/cmd/woa-server/logics/config"
	"hcm/cmd/woa-server/logics/plan"
	rollingserver "hcm/cmd/woa-server/logics/rolling-server"
	"hcm/cmd/woa-server/logics/task/informer"
	"hcm/cmd/woa-server/logics/task/scheduler/record"
	"hcm/cmd/woa-server/logics/task/sops"
	"hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/api/core"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/api-gateway/bkchatapi"
	"hcm/pkg/thirdparty/api-gateway/sopsapi"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/metadata"
	toolsutil "hcm/pkg/tools/util"
	"hcm/pkg/tools/utils/wait"
	"hcm/pkg/tools/uuid"

	"golang.org/x/sync/errgroup"
)

// Matcher matches devices for apply order
type Matcher struct {
	rsLogics     rollingserver.Logics
	planLogics   plan.Logics
	configLogics config.Logics
	informer     informer.Interface
	sops         sopsapi.SopsClientInterface
	sopsOpt      cc.SopsCli
	cc           cmdb.Client
	bkchat       bkchatapi.BkChatClientInterface
	ctx          context.Context
	kt           *kit.Kit
}

// New create a matcher
func New(ctx context.Context, rsLogics rollingserver.Logics, thirdCli *thirdparty.Client, esbCli esb.Client,
	clientConf cc.ClientConfig, informer informer.Interface, planLogics plan.Logics) (*Matcher, error) {

	matcher := &Matcher{
		rsLogics:     rsLogics,
		planLogics:   planLogics,
		configLogics: config.New(thirdCli),
		informer:     informer,
		sops:         thirdCli.Sops,
		sopsOpt:      clientConf.Sops,
		cc:           esbCli.Cmdb(),
		bkchat:       thirdCli.BkChat,
		ctx:          ctx,
		kt:           &kit.Kit{Ctx: ctx, Rid: uuid.UUID()},
	}

	// TODO: get worker num from config
	go matcher.Run(20)

	return matcher, nil
}

// Run starts matcher workers
func (m *Matcher) Run(workers int) {
	for i := 0; i < workers; i++ {
		go wait.Until(m.runWorker, time.Second, m.ctx)
	}

	select {
	case <-m.ctx.Done():
		logs.Infof("matcher exits")
	}
}

// runWorker deals with apply order match task
func (m *Matcher) runWorker() error {
	generateId, err := m.informer.Generate().Pop()
	if err != nil {
		return err
	}

	// get generate record
	generateRecord, err := m.GetGenerateRecord(generateId)
	if err != nil {
		logs.Errorf("failed to get generate record by id: %d, err: %v", generateId, err)
		return err
	}

	// check generate record status
	if generateRecord.Status != types.GenerateStatusSuccess {
		logs.Infof("generate record %d is not done yet, need not match, status: %d", generateId, generateRecord.Status)
		return nil
	}

	// check generate record matched or not
	if generateRecord.IsMatched == true {
		logs.Infof("generate record %d is matched, need not match again", generateId)
		return nil
	}

	// deal match device
	if err := m.matchHandler(generateRecord); err != nil {
		logs.Errorf("failed to match device, order id: %s, err: %v", generateRecord.SubOrderId, err)
		return err
	}

	logs.Infof("match done, generate id: %d, order id: %s", generateId, generateRecord.SubOrderId)

	return nil
}

// FinalApplyStep after deliver device, check order result to regenerate device or reinit
func (m *Matcher) FinalApplyStep(genRecord *types.GenerateRecord, order *types.ApplyOrder) error {
	// set generate record matched
	if err := m.setGenerateRecordMatched(genRecord.GenerateId); err != nil {
		logs.Errorf("failed to update generate record, err: %v, schedule id: %d", err, genRecord.GenerateId)
		return err
	}

	// update apply order status
	if err := m.UpdateApplyOrderStatus(order); err != nil {
		logs.Errorf("failed to update apply order status, order id: %s, err: %v", genRecord.SubOrderId, err)
		return err
	}

	// send ticket done notification
	if err := m.notifyApplyDone(order.OrderId); err != nil {
		logs.Warnf("failed to send apply done notification, order id: %s, err: %v", genRecord.SubOrderId, err)
		return nil
	}
	return nil
}

// matchHandler apply order match handler
func (m *Matcher) matchHandler(genRecord *types.GenerateRecord) error {
	// get apply order by key
	applyOrder, err := m.getApplyOrder(genRecord.SubOrderId)
	if err != nil {
		logs.Errorf("get apply order by key %s failed, err: %v", genRecord.SubOrderId, err)
		return err
	}

	// check order status
	if applyOrder.Status != types.ApplyStatusMatching && applyOrder.Status != types.ApplyStatusGracefulTerminate {
		logs.Infof("apply order %s cannot match for status not Matching, status: %s", genRecord.SubOrderId,
			applyOrder.Status)
		return fmt.Errorf("apply order %s cannot match for status not Matching, status: %s", genRecord.SubOrderId,
			applyOrder.Status)
	}

	// match device
	if err := m.matchDevice(applyOrder, genRecord.GenerateId); err != nil {
		logs.Errorf("failed to match device, order id: %s, err: %v", genRecord.SubOrderId, err)
		return err
	}

	return m.FinalApplyStep(genRecord, applyOrder)
}

// getApplyOrder gets apply order from db by order id
func (m *Matcher) getApplyOrder(orderId string) (*types.ApplyOrder, error) {
	filter := &mapstr.MapStr{
		"suborder_id": orderId,
	}
	order, err := model.Operation().ApplyOrder().GetApplyOrder(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get apply order by id: %s", orderId)
		return nil, err
	}

	return order, nil
}

func (m *Matcher) updateSuspendSteps(order *types.ApplyOrder) error {
	now := time.Now()
	filter := &mapstr.MapStr{
		"suborder_id": order.SubOrderId,
		"step_name":   types.StepNameGenerate,
	}
	doc := &mapstr.MapStr{
		"status":    types.StepStatusFailed,
		"update_at": now,
		"end_at":    now,
		"message":   "can not get generateId, unknown generate status, check YunTi to find if devices are generated",
	}

	if err := model.Operation().ApplyStep().UpdateApplyStep(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update apply 生产 step status to apply status failed, suborderId: %s, err: %v",
			order.SubOrderId, err)
		return err
	}
	return nil
}

func (m *Matcher) updateGenerateFailed(generateId uint64) error {
	filter := &mapstr.MapStr{
		"generate_id": generateId,
	}
	now := time.Now()
	doc := mapstr.MapStr{
		"update_at": now,
		"status":    types.GenerateStatusFailed,
		"message":   "can not get generateId, unknown generate status, check YunTi to find if devices are generated",
	}

	if err := model.Operation().GenerateRecord().UpdateGenerateRecord(context.Background(), filter, &doc); err != nil {
		logs.Errorf("failed to update generate record to failed, generateId: err: %v, %d", err, generateId)
		return err
	}

	return nil
}

// UpdateApplyOrderStatus update apply order status
func (m *Matcher) UpdateApplyOrderStatus(order *types.ApplyOrder) error {
	// 1. get unreleased devices from db
	devices, err := m.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// 2. calculate apply order status by total and matched count
	deviceTypeCountMap := calDeviceTypeCountMap(devices)
	matchedCnt := calMatchCnt(devices)

	genRecords, err := m.GetOrderGenRecords(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get generate records, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	hasGenRecordMatching := false
	isSuspend := false
	suspendCnt := 0

	for _, recordItem := range genRecords {
		if recordItem.Status == types.GenerateStatusInit ||
			recordItem.Status == types.GenerateStatusHandling ||
			recordItem.Status == types.GenerateStatusSuccess && !recordItem.IsMatched {
			hasGenRecordMatching = true
		}

		if recordItem.Status == types.GenerateStatusSuspend {
			isSuspend = true
			suspendCnt += int(recordItem.TotalNum)
			logs.Infof("generate failed, unknown if generate interface was called, task_id not obtained, check machines")
			if err := m.updateGenerateFailed(recordItem.GenerateId); err != nil {
				logs.Errorf("failed to update generate status to failed, suborderId: %s, err: %v", order.SubOrderId,
					err)
			}
		}
	}

	pendingCnt := 0
	status := types.ApplyStatusDone
	stage := types.TicketStageDone
	if matchedCnt < int(order.Total) {
		pendingCnt = int(order.Total) - matchedCnt
		// do not set status to MATCHED_SOME if there are matching tasks
		status = types.ApplyStatusMatchedSome
		if hasGenRecordMatching {
			status = types.ApplyStatusMatching
		}
		stage = types.TicketStageRunning
	}

	if isSuspend && suspendCnt+matchedCnt >= int(order.Total) {
		status = types.ApplyStatusTerminate
		stage = types.TicketStageTerminate
		if err := m.updateSuspendSteps(order); err != nil {
			logs.Errorf("failed to update suspend steps, suborderId: %s, err: %v", order.SubOrderId, err)
		}
	}

	kt := core.NewBackendKit()
	if order.RequireType.IsNeedQuotaManage() {
		appliedTypes := []enumor.AppliedType{enumor.NormalAppliedType, enumor.ResourcePoolAppliedType}

		if err = m.rsLogics.UpdateSubOrderRollingDeliveredCore(kt, order.BkBizId, order.SubOrderId, appliedTypes,
			deviceTypeCountMap); err != nil {
			logs.Errorf("update rolling delivered cpu field failed, err: %v, suborder_id: %s, bizID: %d, "+
				"deviceTypeCountMap: %v, rid: %s", err, order.SubOrderId, order.BkBizId, deviceTypeCountMap, kt.Rid)
			return err
		}
	}

	// 3. do update apply order status
	err = m.updateApplyOrderToDb(kt, order, matchedCnt, pendingCnt, stage, status, deviceTypeCountMap)
	if err != nil {
		return err
	}

	return nil
}

// calMatchCnt calculate matched count
func calMatchCnt(devices []*types.DeviceInfo) int {
	matchedCnt := 0
	for _, device := range devices {
		if !device.IsDelivered {
			continue
		}
		matchedCnt++
	}

	return matchedCnt
}

// calDeviceTypeCountMap calculate matched count
func calDeviceTypeCountMap(devices []*types.DeviceInfo) map[string]int {
	deviceTypeCountMap := make(map[string]int)
	for _, device := range devices {
		if !device.IsDelivered {
			continue
		}

		if _, ok := deviceTypeCountMap[device.DeviceType]; !ok {
			deviceTypeCountMap[device.DeviceType] = 0
		}
		deviceTypeCountMap[device.DeviceType]++
	}
	return deviceTypeCountMap
}

// updateApplyOrderToDb update apply order status to db
func (m *Matcher) updateApplyOrderToDb(kt *kit.Kit, order *types.ApplyOrder, matchedCnt int, pendingCnt int,
	stage types.TicketStage, status types.ApplyStatus, deviceTypeCountMap map[string]int) error {

	filter := &mapstr.MapStr{
		"suborder_id": order.SubOrderId,
	}
	doc := &mapstr.MapStr{
		"success_num": matchedCnt,
		"pending_num": pendingCnt,
		"stage":       stage,
		"status":      status,
		"update_at":   time.Now(),
	}
	if order.ResourceType == types.ResourceTypeCvm {
		sum, err := m.GetCpuCoreSum(kt, deviceTypeCountMap)
		if err != nil {
			logs.Errorf("get cpu core failed, err: %v, deviceTypeCountMap: %v, rid: %s",
				err, deviceTypeCountMap, kt.Rid)
			return err
		}
		doc.Set("delivered_core", sum)

		// 为该子订单匹配CVM资源预测单并生成预测变更记录
		if err = m.planLogics.AddMatchedPlanDemandExpendLogs(kt, order.BkBizId, order); err != nil {
			logs.Errorf("failed to add matched plan demand expend logs, subOrderID: %s, err: %v, subOrder: %+v",
				order.SubOrderId, err, cvt.PtrToVal(order))
			return err
		}
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update apply order, id: %s, err: %v", order.SubOrderId, err)
		return err
	}
	return nil
}

// GetCpuCoreSum 获取机型对应的cpu核数之和
func (m *Matcher) GetCpuCoreSum(kt *kit.Kit, deviceTypeCountMap map[string]int) (int64, error) {
	deviceTypes := make([]string, 0)
	for deviceType := range deviceTypeCountMap {
		deviceTypes = append(deviceTypes, deviceType)
	}
	deviceTypeInfoMap, err := m.configLogics.Device().ListCvmInstanceInfoByDeviceTypes(kt, deviceTypes)
	if err != nil {
		logs.Errorf("get cvm instance info by device type failed, err: %v, device_types: %v, rid: %s",
			err, deviceTypes, kt.Rid)
		return 0, err
	}

	var deliveredCore int64 = 0
	for deviceType, count := range deviceTypeCountMap {
		deviceTypeInfo, ok := deviceTypeInfoMap[deviceType]
		if !ok {
			logs.Errorf("can not find device_type, type: %s, rid: %s", deviceType, kt.Rid)
			return 0, fmt.Errorf("can not find device_type, type: %s", deviceType)
		}
		deliveredCore += deviceTypeInfo.CPUAmount * int64(count)
	}

	return deliveredCore, nil
}

// GetGenerateRecord gets generate record from db by generate id
func (m *Matcher) GetGenerateRecord(id uint64) (*types.GenerateRecord, error) {
	filter := &mapstr.MapStr{
		"generate_id": id,
	}
	recordInfo, err := model.Operation().GenerateRecord().GetGenerateRecord(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get generate record by id: %d", id)
		return nil, err
	}

	return recordInfo, nil
}

// GetOrderGenRecords gets all generate records related to given order
func (m *Matcher) GetOrderGenRecords(suborderId string) ([]*types.GenerateRecord, error) {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: pkg.BKNoLimit,
	}

	records, err := model.Operation().GenerateRecord().FindManyGenerateRecord(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get generate record by order id: %s", suborderId)
		return nil, err
	}

	return records, nil
}

// setGenerateRecordMatched set generate record matched
func (m *Matcher) setGenerateRecordMatched(generateId uint64) error {
	filter := &mapstr.MapStr{
		"generate_id": generateId,
	}

	doc := mapstr.MapStr{
		"is_matched": true,
		"update_at":  time.Now(),
	}

	if err := model.Operation().GenerateRecord().UpdateGenerateRecord(context.Background(), filter, &doc); err != nil {
		logs.Errorf("failed to update generate record, generate id: %d, update: %+v, err: %v", generateId, doc,
			err)
		return err
	}

	return nil
}

// InitDevices start init devices
func (m *Matcher) InitDevices(order *types.ApplyOrder, unreleased []*types.DeviceInfo) ([]*types.DeviceInfo, error) {
	// start init step
	if err := record.StartStep(order.SubOrderId, types.StepNameInit); err != nil {
		logs.Errorf("failed to start init step, order id: %s, err: %v", order.SubOrderId, err)
		return nil, err
	}

	successDeviceMap, errMap := m.ProcessInitStep(unreleased)
	if len(errMap) > 0 {
		// todo 暂时和原逻辑保持一致，这里err不做处理，ProcessInitStep内已经有打印错误日志
	}

	// update init step
	if err := record.UpdateInitStep(order.SubOrderId, order.Total); err != nil {
		logs.Errorf("failed to update init step, subOrderID: %s, err: %v", order.SubOrderId, err)
		return nil, err
	}

	return maps.Values(successDeviceMap), nil
}

// DeliverDevices deliver devices to business
func (m *Matcher) DeliverDevices(order *types.ApplyOrder, observeDevices []*types.DeviceInfo) error {
	// start deliver step
	if err := record.StartStep(order.SubOrderId, types.StepNameDeliver); err != nil {
		logs.Errorf("failed to start deliver step, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// deliver devices to business
	// TODO: batch processing
	for _, device := range observeDevices {
		if err := m.DeliverDevice(device, order); err != nil {
			logs.Errorf("failed to deliver device, subOrderId: %s, ip: %s, err: %v", order.SubOrderId, device.Ip, err)
			continue
		}
	}

	// update deliver step
	if err := record.UpdateDeliverStep(order.SubOrderId, order.Total); err != nil {
		logs.Errorf("failed to update init step, subOrderId: %s, err: %v", order.SubOrderId, err)
		return err
	}
	return nil
}

// ProcessInitStep process init step
func (m *Matcher) ProcessInitStep(devices []*types.DeviceInfo) (map[int]*types.DeviceInfo, map[int]error) {
	maxRetry := 3
	errMap := make(map[int]error)
	deviceInitMsgMap := make(map[int]*types.DeviceInitMsg)
	eg := errgroup.Group{}
	eg.SetLimit(10)
	var lock sync.Mutex

	// 1. 创建主机初始化任务
	for idx, device := range devices {
		curDevice := device
		curIdx := idx
		eg.Go(func() error {
			var err error
			var initMsg *types.DeviceInitMsg
			for try := 0; try < maxRetry; try++ {
				if initMsg, err = m.initDevice(curDevice); err != nil {
					logs.Errorf("failed to init device, will retry in 60s, ip: %s, err: %v", curDevice.Ip, err)
					time.Sleep(time.Second)
					continue
				}
				break
			}
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				errMap[curIdx] = err
				return nil
			}
			deviceInitMsgMap[curIdx] = initMsg
			return nil
		})
	}
	_ = eg.Wait()

	// 2. 检查主机初始化任务是否执行完成
	successDeviceMap := make(map[int]*types.DeviceInfo)
	for idx, msg := range deviceInitMsgMap {
		curMsg := msg
		curIdx := idx
		eg.Go(func() error {
			err := m.CheckSopsUpdate(curMsg.BizID, curMsg.Device, curMsg.JobUrl, curMsg.JobID)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				logs.Errorf("failed to check sops task, ip: %s, err: %v", curMsg.Device.Ip, err)
				errMap[curIdx] = err
				return nil
			}
			successDeviceMap[curIdx] = curMsg.Device
			return nil
		})
	}
	_ = eg.Wait()

	return successDeviceMap, errMap
}

// matchDevice deal match device tasks
func (m *Matcher) matchDevice(order *types.ApplyOrder, genId uint64) error {
	// 1. get unreleased devices from db
	unreleased, err := m.getGeneratedDevice(genId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	observeDevices, err := m.InitDevices(order, unreleased)

	if order.EnableDiskCheck {
		observeDevices, err = m.RunDiskCheck(order, observeDevices)
		if err != nil {
			logs.Errorf("failed to run disk check task, order id: %s, err: %v", order.SubOrderId, err)
			return err
		}
	}

	return m.DeliverDevices(order, observeDevices)
}

// getGeneratedDevice gets generated devices bindings to generate record
func (m *Matcher) getGeneratedDevice(genId uint64) ([]*types.DeviceInfo, error) {
	filter := &mapstr.MapStr{
		"generate_id": genId,
	}

	devices, err := model.Operation().DeviceInfo().GetDeviceInfo(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get binding devices to generate id %d, err: %v", genId, err)
		return nil, err
	}

	return devices, nil
}

// getUnreleasedDevice gets unreleased devices bindings to current apply order
func (m *Matcher) getUnreleasedDevice(orderId string) ([]*types.DeviceInfo, error) {
	filter := &mapstr.MapStr{
		"suborder_id": orderId,
	}

	devices, err := model.Operation().DeviceInfo().GetDeviceInfo(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get binding devices to order %s, err: %v", orderId, err)
		return nil, err
	}

	return devices, nil
}

// initDevice executes device initialization task
func (m *Matcher) initDevice(info *types.DeviceInfo) (*types.DeviceInitMsg, error) {
	if info.IsInited {
		logs.Infof("host %s is initialized, need not init", info.Ip)
		return nil, nil
	}

	// create init record
	if err := record.CreateInitRecord(info.SubOrderId, info.Ip); err != nil {
		logs.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
		return nil, fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
	}

	// 1. create job
	// 根据IP获取主机信息
	hostInfo, err := m.cc.GetHostInfoByIP(m.kt.Ctx, m.kt.Header(), info.Ip, 0)
	if err != nil {
		logs.Errorf("sops:process:check:matcher:ieod init, get host info by ip failed, ip: %s, infoBkBizID: %d, "+
			"err: %v", info.Ip, info.BkBizId, err)
		return nil, err
	}

	// 根据bkHostID去cmdb获取bkBizID
	bkBizIDs, err := m.cc.GetHostBizIds(m.kt.Ctx, m.kt.Header(), []int64{hostInfo.BkHostID})
	if err != nil {
		logs.Errorf("sops:process:check:matcher:ieod init, get host info by host id failed, ip: %s, infoBkBizID: %d, "+
			"bkHostID: %d, err: %v", info.Ip, info.BkBizId, hostInfo.BkHostID, err)
		return nil, err
	}
	bkBizID, ok := bkBizIDs[hostInfo.BkHostID]
	if !ok {
		logs.Errorf("can not find biz id by host id: %d", hostInfo.BkHostID)
		return nil, fmt.Errorf("can not find biz id by host id: %d", hostInfo.BkHostID)
	}
	jobId, jobUrl, err := sops.CreateInitSopsTask(m.kt, m.sops, info.Ip, m.sopsOpt.DevnetIP, bkBizID, hostInfo.BkOsType,
		info.SubOrderId)
	if err != nil {
		logs.Errorf("sops:process:check:matcher:ieod init device, host %s failed to initialize, infoBkBizID: %d, "+
			"bkBizID: %d, bkHostID: %d, err: %v", info.Ip, info.BkBizId, bkBizID, info.BkHostId, err)
		// update init record
		errRecord := record.UpdateInitRecord(info.SubOrderId, info.Ip, "", "", err.Error(), types.InitStatusFailed)
		if errRecord != nil {
			logs.Errorf("update init record failed, host ip: %s, bkBidID: %d, bkHostID: %d, err: %v",
				info.Ip, info.BkBizId, info.BkHostId, errRecord)
			return nil, fmt.Errorf("update init record failed, host ip: %s, err: %v", info.Ip, errRecord)
		}
		return nil, fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
	}

	jobIDStr := strconv.FormatInt(jobId, 10)
	// update init record
	errRecord := record.UpdateInitRecord(info.SubOrderId, info.Ip, jobIDStr, jobUrl, "handling",
		types.InitStatusHandling)
	if errRecord != nil {
		logs.Warnf("host %s failed to update initialize record, jobID: %d, jobUrl: %s, bkBizID: %s, err: %v",
			info.Ip, jobId, jobUrl, bkBizID, errRecord)
	}

	return &types.DeviceInitMsg{Device: info, JobUrl: jobUrl, JobID: jobIDStr, BizID: bkBizID}, nil
}

// CheckSopsUpdate 检查sops任务状态并更新
func (m *Matcher) CheckSopsUpdate(bkBizID int64, info *types.DeviceInfo, jobUrl string, jobIDStr string) error {
	// 1. get job status
	jobId, err := strconv.ParseInt(jobIDStr, 10, 64)
	if err != nil {
		logs.Errorf("can not get jobId by jobIDStr, jobIDStr: %s, err: %v", jobIDStr, err)
		return fmt.Errorf("can not get jobId by jobIDStr, jobIDStr: %s", jobIDStr)
	}

	if _, err = sops.CheckTaskStatus(m.kt, m.sops, jobId, bkBizID); err != nil {
		logs.Infof("sops:process:check:matcher:ieod init device, host %s failed to initialize, jobID: %d, "+
			"jobUrl: %s, bkBizID: %d, err: %v", info.Ip, jobId, jobUrl, bkBizID, err)
		// update init record
		errRecord := record.UpdateInitRecord(info.SubOrderId, info.Ip, jobIDStr, jobUrl,
			err.Error(), types.InitStatusFailed)
		if errRecord != nil {
			logs.Errorf("host %s failed to initialize, bkBizID: %d, jobID: %d, jobUrl: %s, err: %v",
				info.Ip, bkBizID, jobId, jobUrl, errRecord)
			return fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, errRecord)
		}
		return fmt.Errorf("host %s failed to initialize, jobID: %d, err: %v", info.Ip, jobId, err)
	}

	// 2. update device status
	info.InitTaskId = strconv.FormatInt(jobId, 10)
	info.InitTaskLink = jobUrl
	if err := m.SetDeviceInited(info); err != nil {
		logs.Errorf("host %s failed to initialize, jobID: %d, jobUrl: %s, err: %v", info.Ip, jobId, jobUrl, err)
		return fmt.Errorf("host %s failed to initialize, jobID: %d, jobUrl: %s, err: %v", info.Ip, jobId, jobUrl, err)
	}

	// update init record
	if err := record.UpdateInitRecord(info.SubOrderId, info.Ip, jobIDStr, jobUrl, "success",
		types.InitStatusSuccess); err != nil {
		logs.Errorf("host %s failed to initialize, bkBizID: %d, jobId: %d, jobUrl: %s, err: %v",
			info.Ip, bkBizID, jobId, jobUrl, err)
		return fmt.Errorf("host %s failed to initialize, jobID: %d, jobUrl: %s, err: %v", info.Ip, jobId, jobUrl, err)
	}
	return nil
}

// checkDeviceDisk executes device disk check task
func (m *Matcher) checkDeviceDisk(info *types.DeviceInfo) error {
	if info.IsDiskChecked {
		logs.Infof("host %s is disk-checked, need not disk check", info.Ip)
		return nil
	}

	return nil
}

// DeliverDevice delivers device to business
func (m *Matcher) DeliverDevice(info *types.DeviceInfo, order *types.ApplyOrder) error {
	if info.IsDelivered {
		logs.Infof("host %s is delivered, need not deliver, subOrderID: %s", info.Ip, order.SubOrderId)
		return nil
	}

	// create deliver record
	if err := record.CreateDeliverRecord(info); err != nil {
		logs.Errorf("failed to deliver device, ip: %s, subOrderID: %s, err: %v", info.Ip, order.SubOrderId, err)
		return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
	}
	// 1. set host module and host operator
	if err := m.transferHostAndSetOperator(info, order); err != nil {
		logs.Errorf("failed to deliver device, ip: %s, subOrderID: %s, err: %v", info.Ip, order.SubOrderId, err)
		// update deliver record
		if errRecord := record.UpdateDeliverRecord(info, err.Error(), types.DeliverStatusFailed); errRecord != nil {
			logs.Errorf("failed to deliver device, ip: %s, subOrderID: %s, err: %v", info.Ip, order.SubOrderId, err)
			return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
		}
		return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
	}
	// 2. update device status
	if err := m.SetDeviceDelivered(info); err != nil {
		logs.Errorf("failed to deliver device, ip: %s, subOrderID: %s, err: %v", info.Ip, order.SubOrderId, err)
		return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
	}

	// update deliver record
	if err := record.UpdateDeliverRecord(info, "success", types.DeliverStatusSuccess); err != nil {
		logs.Errorf("failed to deliver device, ip: %s, subOrderID: %s, err: %v", info.Ip, order.SubOrderId, err)
		return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
	}

	return nil
}

// setDeviceChecked set device checked flag
func (m *Matcher) setDeviceChecked(info *types.DeviceInfo) error {
	filter := &mapstr.MapStr{
		"suborder_id": info.SubOrderId,
		"ip":          info.Ip,
	}
	doc := &mapstr.MapStr{
		"is_checked": true,
		"update_at":  time.Now(),
	}

	if err := model.Operation().DeviceInfo().UpdateDeviceInfo(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update device checked flag, ip: %s, err: %v", info.Ip, err)
		return err
	}

	info.IsChecked = true

	return nil
}

// SetDeviceInited set device inited flag
func (m *Matcher) SetDeviceInited(info *types.DeviceInfo) error {
	filter := &mapstr.MapStr{
		"suborder_id": info.SubOrderId,
		"ip":          info.Ip,
	}
	doc := &mapstr.MapStr{
		"is_inited":      true,
		"init_task_id":   info.InitTaskId,
		"init_task_link": info.InitTaskLink,
		"update_at":      time.Now(),
	}

	if err := model.Operation().DeviceInfo().UpdateDeviceInfo(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update device inited flag, ip: %s, err: %v", info.Ip, err)
		return err
	}

	info.IsInited = true

	return nil
}

// setDeviceDiskChecked set device disk-checked flag
func (m *Matcher) setDeviceDiskChecked(info *types.DeviceInfo) error {
	filter := &mapstr.MapStr{
		"suborder_id": info.SubOrderId,
		"ip":          info.Ip,
	}
	doc := &mapstr.MapStr{
		"is_disk_checked":      true,
		"disk_check_task_id":   info.InitTaskId,
		"disk_check_task_link": info.InitTaskLink,
		"update_at":            time.Now(),
	}

	if err := model.Operation().DeviceInfo().UpdateDeviceInfo(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update device disk-checked flag, ip: %s, err: %v", info.Ip, err)
		return err
	}

	info.IsInited = true

	return nil
}

// SetDeviceDelivered set device delivered flag
func (m *Matcher) SetDeviceDelivered(info *types.DeviceInfo) error {
	filter := &mapstr.MapStr{
		"suborder_id": info.SubOrderId,
		"ip":          info.Ip,
	}
	doc := &mapstr.MapStr{
		"is_delivered": true,
		"update_at":    time.Now(),
	}

	if err := model.Operation().DeviceInfo().UpdateDeviceInfo(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update device delivered flag, ip: %s, err: %v", info.Ip, err)
		return err
	}

	return nil
}

func (m *Matcher) notifyApplyDone(orderId uint64) error {
	// check if all apply suborders done
	filter := map[string]interface{}{
		"order_id": orderId,
		"status": map[string]interface{}{
			pkg.BKDBNE: types.ApplyStatusDone,
		},
	}

	cnt, err := model.Operation().ApplyOrder().CountApplyOrder(context.Background(), filter)
	if err != nil {
		return err
	}
	if cnt > 0 {
		// exist suborder not done, need not notify
		return nil
	}

	filterTicket := &mapstr.MapStr{
		"order_id": orderId,
	}

	ticket, err := model.Operation().ApplyTicket().GetApplyTicket(context.Background(), filterTicket)
	if err != nil {
		return nil
	}

	// TODO: add verification after front end set enable notice by default
	/*
		if !ticket.EnableNotice {
			// need not notify
			return nil
		}
	*/

	users := []string{ticket.User}
	users = append(users, ticket.Follower...)
	users = toolsutil.StrArrayUnique(users)
	noticeFmt := m.bkchat.GetNoticeFmt()
	bizName := m.getBizName(ticket.BkBizId)
	requireName := ticket.RequireType.GetName()
	createTime := ticket.CreateAt.Local().Format(constant.DateTimeLayout)
	if ticket.CreateAt.Location() == time.UTC {
		location, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			logs.Warnf("scheduler:logics:bkchat:notifyApplyDone:failed, orderId: %d, err: %v, createAt: %+v",
				orderId, err, ticket.CreateAt)
			return err
		}
		createTime = ticket.CreateAt.In(location).Format(constant.DateTimeLayout)
	}
	content := fmt.Sprintf(noticeFmt, orderId, orderId, ticket.User, bizName, requireName, createTime, ticket.Remark,
		orderId, ticket.BkBizId)

	for _, user := range users {
		resp, err := m.bkchat.SendApplyDoneMsg(nil, nil, user, content)
		if err != nil {
			logs.Warnf("scheduler:logics:bkchat:notifyApplyDone:failed, failed to send bkchat message, err: %v", err)
			continue
		}
		if resp.Code != 0 {
			logs.Warnf("scheduler:logics:bkchat:notifyApplyDone:failed, failed to send bkchat message, "+
				"code: %d, msg: %s", resp.Code, resp.Msg)
			continue
		}
	}

	return nil
}

// RunDiskCheck 执行磁盘检查
func (m *Matcher) RunDiskCheck(order *types.ApplyOrder, devices []*types.DeviceInfo) ([]*types.DeviceInfo, error) {
	// start init step
	if err := record.StartStep(order.SubOrderId, types.StepNameDiskCheck); err != nil {
		logs.Errorf("failed to start init step, order id: %s, err: %v", order.SubOrderId, err)
		return nil, err
	}

	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	observeDevices := make([]*types.DeviceInfo, 0)
	appendError := func(err error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, err)
	}
	appendDevice := func(device *types.DeviceInfo) {
		mutex.Lock()
		defer mutex.Unlock()
		observeDevices = append(observeDevices, device)
	}
	for _, device := range devices {
		wg.Add(1)
		go func(device *types.DeviceInfo) {
			defer wg.Done()

			// check device disk
			maxRetry := 3
			var err error = nil
			for try := 0; try < maxRetry; try++ {
				if err = m.checkDeviceDisk(device); err != nil {
					logs.Errorf("failed to check device disk, will retry in 60s, ip: %s, err: %v", device.Ip, err)
					time.Sleep(180 * time.Second)
					continue
				}
				break
			}

			if err != nil {
				appendError(err)
			} else {
				appendDevice(device)
			}
		}(device)
	}
	wg.Wait()

	// update disk check step
	if err := record.UpdateDiskCheckStep(order.SubOrderId, order.Total); err != nil {
		logs.Errorf("failed to update init step, order id: %s, err: %v", order.SubOrderId, err)
		return nil, err
	}

	return observeDevices, nil
}
