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

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	commonutil "hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/common/utils/wait"
	"hcm/cmd/woa-server/logics/task/informer"
	"hcm/cmd/woa-server/logics/task/scheduler/record"
	"hcm/cmd/woa-server/model/task"
	"hcm/cmd/woa-server/thirdparty"
	"hcm/cmd/woa-server/thirdparty/bkchatapi"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	"hcm/cmd/woa-server/thirdparty/sojobapi"
	"hcm/cmd/woa-server/thirdparty/sopsapi"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/logs"
)

// Matcher matches devices for apply order
type Matcher struct {
	informer informer.Interface
	sojob    sojobapi.SojobClientInterface
	sops     sopsapi.SopsClientInterface
	cc       cmdb.Client
	bkchat   bkchatapi.BkChatClientInterface
	ctx      context.Context
}

// New create a matcher
func New(ctx context.Context, thirdCli *thirdparty.Client, esbCli esb.Client, informer informer.Interface) (
	*Matcher, error) {

	matcher := &Matcher{
		informer: informer,
		sojob:    thirdCli.SoJob,
		sops:     thirdCli.Sops,
		cc:       esbCli.Cmdb(),
		bkchat:   thirdCli.BkChat,
		ctx:      ctx,
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
	generateRecord, err := m.getGenerateRecord(generateId)
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

// matchHandler apply order match handler
func (m *Matcher) matchHandler(genRecord *types.GenerateRecord) error {
	// get apply order by key
	applyOrder, err := m.getApplyOrder(genRecord.SubOrderId)
	if err != nil {
		logs.Errorf("get apply order by key %s failed, err: %v", genRecord.SubOrderId, err)
		return err
	}

	// check order status
	if applyOrder.Status != types.ApplyStatusMatching {
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

	// set generate record matched
	if err := m.setGenerateRecordMatched(genRecord.GenerateId); err != nil {
		logs.Errorf("failed to update generate record, schedule id: %d, err: %v", genRecord.GenerateId, err)
		return err
	}

	// update apply order status
	if err := m.updateApplyOrderStatus(applyOrder); err != nil {
		logs.Errorf("failed to update apply order status, order id: %s, err: %v", genRecord.SubOrderId, err)
		return err
	}

	// send ticket done notification
	if err := m.notifyApplyDone(applyOrder.OrderId); err != nil {
		logs.Warnf("failed to send apply done notification, order id: %s, err: %v", genRecord.SubOrderId, err)
		return nil
	}

	return nil
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

// updateApplyOrderStatus update apply order status
func (m *Matcher) updateApplyOrderStatus(order *types.ApplyOrder) error {
	// 1. get unreleased devices from db
	devices, err := m.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// 2. calculate apply order status by total and matched count
	matchedCnt := 0
	for _, device := range devices {
		if device.IsDelivered {
			matchedCnt++
		}
	}

	genRecords, err := m.getOrderGenRecords(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get generate records, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	hasGenRecordMatching := false
	for _, record := range genRecords {
		if record.Status == types.GenerateStatusInit ||
			record.Status == types.GenerateStatusHandling ||
			record.Status == types.GenerateStatusSuccess && record.IsMatched == false {
			hasGenRecordMatching = true
			break
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

	// 3. do update apply order status
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
	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update apply order, id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	return nil
}

// getGenerateRecord gets generate record from db by generate id
func (m *Matcher) getGenerateRecord(id uint64) (*types.GenerateRecord, error) {
	filter := &mapstr.MapStr{
		"generate_id": id,
	}
	record, err := model.Operation().GenerateRecord().GetGenerateRecord(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get generate record by id: %d", id)
		return nil, err
	}

	return record, nil
}

// getOrderGenRecords gets all generate records related to given order
func (m *Matcher) getOrderGenRecords(suborderId string) ([]*types.GenerateRecord, error) {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: common.BKNoLimit,
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

// matchDevice deal match device tasks
func (m *Matcher) matchDevice(order *types.ApplyOrder, genId uint64) error {
	// 1. get unreleased devices from db
	unreleased, err := m.getGeneratedDevice(genId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// start init step
	if err := record.StartStep(order.SubOrderId, types.StepNameInit); err != nil {
		logs.Errorf("failed to start init step, order id: %s, err: %v", order.SubOrderId, err)
		return err
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
	for _, device := range unreleased {
		wg.Add(1)
		go func(device *types.DeviceInfo) {
			defer wg.Done()
			// TODO: check device not support yet
			// 2. check devices
			/*
				if err := m.checkDevice(device); err != nil {
					logs.Errorf("failed to check device, ip: %s, err: %v", device.IP, err)
					continue
				}*/

			// 3. init devices
			maxRetry := 3
			var err error = nil
			for try := 0; try < maxRetry; try++ {
				if err = m.initDevice(device); err != nil {
					logs.Errorf("failed to init device, will retry in 60s, ip: %s, err: %v", device.Ip, err)
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

	// update init step
	if err := record.UpdateInitStep(order.SubOrderId, order.Total); err != nil {
		logs.Errorf("failed to update init step, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	if order.EnableDiskCheck {
		observeDevices, err = m.runDiskCheck(order, observeDevices)
		if err != nil {
			logs.Errorf("failed to run disk check task, order id: %s, err: %v", order.SubOrderId, err)
			return err
		}
	}

	// start deliver step
	if err := record.StartStep(order.SubOrderId, types.StepNameDeliver); err != nil {
		logs.Errorf("failed to start deliver step, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// 4. deliver devices to business
	// TODO: batch processing
	for _, device := range observeDevices {
		if err := m.deliverDevice(device, order); err != nil {
			logs.Errorf("failed to deliver device, ip: %s, err: %v", device.Ip, err)
			continue
		}
	}

	// update deliver step
	if err := record.UpdateDeliverStep(order.SubOrderId, order.Total); err != nil {
		logs.Errorf("failed to update init step, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	return nil
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

// checkDevice executes device quality check task
func (m *Matcher) checkDevice(info *types.DeviceInfo) error {
	if info.IsChecked {
		logs.Infof("host %s is checked, need not check", info.Ip)
		return nil
	}

	// 1. create job
	ips := []string{info.Ip}
	// TODO: ieod_check not support yet
	jobId, err := m.createSoJob("ieod_check", ips)
	if err != nil {
		logs.Errorf("host %s failed to check, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to check, err: %v", info.Ip, err)
	}

	// 2. get job status
	if err := m.checkJobStatus(jobId); err != nil {
		logs.Errorf("host %s failed to check, job id: %d, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to check, job id: %d, err: %v", info.Ip, jobId, err)
	}

	// 3. update device status
	if err := m.setDeviceChecked(info); err != nil {
		logs.Errorf("host %s failed to check, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to check, err: %v", info.Ip, err)
	}

	return nil
}

// initDevice executes device initialization task
func (m *Matcher) initDevice(info *types.DeviceInfo) error {
	if info.IsInited {
		logs.Infof("host %s is initialized, need not init", info.Ip)
		return nil
	}

	// create init record
	if err := record.CreateInitRecord(info.SubOrderId, info.Ip); err != nil {
		logs.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
	}

	// 1. create job
	ips := []string{info.Ip}
	jobId, err := m.createSoJob("ieod_init", ips)
	if err != nil {
		logs.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
		// update init record
		errRecord := record.UpdateInitRecord(info.SubOrderId, info.Ip, "", err.Error(), types.InitStatusFailed)
		if errRecord != nil {
			logs.Errorf("host %s failed to initialize, err: %v", info.Ip, errRecord)
			return fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, errRecord)
		}
		return fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
	}

	// update init record
	errRecord := record.UpdateInitRecord(info.SubOrderId, info.Ip, strconv.Itoa(jobId), "handling",
		types.InitStatusHandling)
	if errRecord != nil {
		logs.Warnf("host %s failed to update initialize record, err: %v", info.Ip, errRecord)
	}

	// 2. get job status
	if err := m.checkJobStatus(jobId); err != nil {
		logs.Infof("host %s failed to initialize, job id: %d, err: %v", info.Ip, jobId, err)
		// update init record
		errRecord := record.UpdateInitRecord(info.SubOrderId, info.Ip, strconv.Itoa(jobId), err.Error(),
			types.InitStatusFailed)
		if errRecord != nil {
			logs.Errorf("host %s failed to initialize, err: %v", info.Ip, errRecord)
			return fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, errRecord)
		}
		return fmt.Errorf("host %s failed to initialize, job id: %d, err: %v", info.Ip, jobId, err)
	}

	// 3. update device status
	info.InitTaskId = strconv.Itoa(jobId)
	info.InitTaskLink = sojobapi.InitTaskLinkPrefix + info.InitTaskId
	if err := m.setDeviceInited(info); err != nil {
		logs.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
	}

	// update init record
	if err := record.UpdateInitRecord(info.SubOrderId, info.Ip, strconv.Itoa(jobId), "success",
		types.InitStatusSuccess); err != nil {
		logs.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to initialize, err: %v", info.Ip, err)
	}

	return nil
}

// checkDeviceDisk executes device disk check task
func (m *Matcher) checkDeviceDisk(info *types.DeviceInfo) error {
	if info.IsDiskChecked {
		logs.Infof("host %s is disk-checked, need not disk check", info.Ip)
		return nil
	}

	// create disk check record
	if err := record.CreateDiskCheckRecord(info.SubOrderId, info.Ip); err != nil {
		logs.Errorf("host %s failed to disk check, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to disk check, err: %v", info.Ip, err)
	}

	// 1. create job
	ips := []string{info.Ip}

	jobId, err := m.createSoJob("benchmark", ips)
	if err != nil {
		logs.Errorf("host %s failed to disk check, err: %v", info.Ip, err)
		// update init record
		errRecord := record.UpdateDiskCheckRecord(info.SubOrderId, info.Ip, "", err.Error(),
			types.DiskCheckStatusFailed)
		if errRecord != nil {
			logs.Errorf("host %s failed to disk check, err: %v", info.Ip, errRecord)
			return fmt.Errorf("host %s failed to disk check, err: %v", info.Ip, errRecord)
		}
		return fmt.Errorf("host %s failed to disk check, err: %v", info.Ip, err)
	}

	// update init record
	errRecord := record.UpdateDiskCheckRecord(info.SubOrderId, info.Ip, strconv.Itoa(jobId), "handling",
		types.DiskCheckStatusHandling)
	if errRecord != nil {
		logs.Warnf("host %s failed to update disk check record, err: %v", info.Ip, errRecord)
	}

	// 2. get job status
	if err := m.checkJobStatusDetail(jobId); err != nil {
		logs.Infof("host %s failed to disk check, job id: %d, err: %v", info.Ip, jobId, err)
		// update init record
		errRecord := record.UpdateDiskCheckRecord(info.SubOrderId, info.Ip, strconv.Itoa(jobId), err.Error(),
			types.DiskCheckStatusFailed)
		if errRecord != nil {
			logs.Errorf("host %s failed to disk check, err: %v", info.Ip, errRecord)
			return fmt.Errorf("host %s failed to disk check, err: %v", info.Ip, errRecord)
		}
		return fmt.Errorf("host %s failed to disk check, job id: %d, err: %v", info.Ip, jobId, err)
	}

	// 3. update device status
	info.DiskCheckTaskId = strconv.Itoa(jobId)
	info.DiskCheckTaskLink = sojobapi.InitTaskLinkPrefix + info.DiskCheckTaskId
	if err := m.setDeviceDiskChecked(info); err != nil {
		logs.Errorf("host %s failed to disk check, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to disk check, err: %v", info.Ip, err)
	}

	// update disk check record
	if err := record.UpdateDiskCheckRecord(info.SubOrderId, info.Ip, strconv.Itoa(jobId), "success",
		types.DiskCheckStatusSuccess); err != nil {
		logs.Errorf("host %s failed to disk check, err: %v", info.Ip, err)
		return fmt.Errorf("host %s failed to disk check, err: %v", info.Ip, err)
	}

	return nil
}

// deliverDevice delivers device to business
func (m *Matcher) deliverDevice(info *types.DeviceInfo, order *types.ApplyOrder) error {
	if info.IsDelivered {
		logs.Infof("host %s is delivered, need not deliver", info.Ip)
		return nil
	}

	// create deliver record
	if err := record.CreateDeliverRecord(info); err != nil {
		logs.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
		return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
	}

	// 1. set host module and host operator
	if err := m.transferHostAndSetOperator(info, order); err != nil {
		logs.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
		// update deliver record
		if errRecord := record.UpdateDeliverRecord(info, err.Error(), types.DeliverStatusFailed); errRecord != nil {
			logs.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
			return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
		}
		return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
	}

	// 2. update device status
	if err := m.setDeviceDelivered(info); err != nil {
		logs.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
		return fmt.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
	}

	// update deliver record
	if err := record.UpdateDeliverRecord(info, "success", types.DeliverStatusSuccess); err != nil {
		logs.Errorf("failed to deliver device, ip: %s, err: %v", info.Ip, err)
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

// setDeviceInited set device inited flag
func (m *Matcher) setDeviceInited(info *types.DeviceInfo) error {
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

// setDeviceDelivered set device delivered flag
func (m *Matcher) setDeviceDelivered(info *types.DeviceInfo) error {
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
			common.BKDBNE: types.ApplyStatusDone,
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
	users = commonutil.StrArrayUnique(users)
	noticeFmt := m.bkchat.GetNoticeFmt()
	bizName := m.getBizName(ticket.BkBizId)
	requireName := m.getRequireName(ticket.RequireType)
	createTime := ticket.CreateAt.Local().Format("2006-01-02 15:04:05")
	content := fmt.Sprintf(noticeFmt, orderId, orderId, ticket.User, bizName, requireName, createTime, ticket.Remark,
		orderId)

	for _, user := range users {
		resp, err := m.bkchat.SendApplyDoneMsg(nil, nil, user, content)
		if err != nil {
			logs.Warnf("failed to send bkchat message, err: %v", err)
			continue
		}
		if resp.Code != 0 {
			logs.Warnf("failed to send bkchat message, code: %d, msg: %s", resp.Code, resp.Msg)
			continue
		}
	}

	return nil
}

func (m *Matcher) getRequireName(requireType int64) string {
	switch requireType {
	case 1:
		return "常规项目"
	case 2:
		return "春节保障"
	case 3:
		return "机房裁撤"
	case 4:
		return "故障替换"
	default:
		return ""
	}
}

func (m *Matcher) runDiskCheck(order *types.ApplyOrder, devices []*types.DeviceInfo) ([]*types.DeviceInfo, error) {
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
