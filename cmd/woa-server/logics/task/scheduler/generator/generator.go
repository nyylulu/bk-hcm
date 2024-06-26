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

// Package generator provides ...
package generator

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/querybuilder"
	utils "hcm/cmd/woa-server/common/util"
	"hcm/cmd/woa-server/logics/config"
	poolLogics "hcm/cmd/woa-server/logics/pool"
	"hcm/cmd/woa-server/logics/task/scheduler/algorithm"
	"hcm/cmd/woa-server/model/task"
	"hcm/cmd/woa-server/thirdparty"
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	"hcm/cmd/woa-server/thirdparty/dvmapi"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/cmd/woa-server/thirdparty/esb/cmdb"
	cfgtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Generator generates vm devices
type Generator struct {
	cvm          cvmapi.CVMClientInterface
	dvm          dvmapi.DVMClientInterface
	cc           cmdb.Client
	ctx          context.Context
	configLogics config.Logics
	poolLogics   poolLogics.Logics
	clientConf   cc.ClientConfig

	predicateFuncs map[string]algorithm.FitPredicate
	priorityFuncs  []algorithm.PriorityConfig
}

// New creates a generator
func New(ctx context.Context, thirdCli *thirdparty.Client, esbCli esb.Client, clientConf cc.ClientConfig) (
	*Generator, error) {

	predicateFuncs := initPredicateFuncs()
	priorityFuncs := initpriorityFuncs()

	generator := &Generator{
		cvm:            thirdCli.CVM,
		dvm:            thirdCli.DVM,
		cc:             esbCli.Cmdb(),
		predicateFuncs: predicateFuncs,
		priorityFuncs:  priorityFuncs,
		ctx:            ctx,
		clientConf:     clientConf,
		configLogics:   config.New(thirdCli),
		poolLogics:     poolLogics.New(ctx, clientConf, thirdCli, esbCli),
	}

	return generator, nil
}

func initPredicateFuncs() map[string]algorithm.FitPredicate {
	predicateFuncs := map[string]algorithm.FitPredicate{
		"VMFitHostVirtualRatio": algorithm.VMFitHostVirtualRatio,
		"VMFitRegion":           algorithm.VMFitRegion,
		"VMFitCampus":           algorithm.VMFitCampus,
		"VMFitKernel":           algorithm.VMFitKernel,
		"VMFitCpuProvider":      algorithm.VMFitCpuProvider,
	}

	return predicateFuncs
}

func initpriorityFuncs() []algorithm.PriorityConfig {
	priorityFuncs := []algorithm.PriorityConfig{
		{
			Function: algorithm.CalculateBalancedResourceAllocation,
			Weight:   10,
		},
	}

	return priorityFuncs
}

// GenerateCVM generates cvm devices
func (g *Generator) GenerateCVM(kt *kit.Kit, order *types.ApplyOrder) error {
	// 1. get history generated devices
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// check if need generate cvm
	existCount := uint(len(existDevices))
	if existCount >= order.Total {
		logs.Infof("apply order %s has been scheduled %d cvm", order.SubOrderId, existCount)
		// check if need retry match task
		if err := g.retryMatchDevice(existDevices); err != nil {
			logs.Warnf("failed to retry match device, order id: %s, err: %v", order.SubOrderId, err)
		}
		return nil
	}

	logs.Infof("apply order %s existing device number: %d", order.SubOrderId, existCount)

	// for given zone case
	if order.Spec.Zone != "" && order.Spec.Zone != cvmapi.CvmSeparateCampus {
		if err := g.generateCVMConcentrate(kt, order, existDevices); err != nil {
			logs.Errorf("failed to generate cvm in zone %s, suborder id: %s", order.Spec.Zone, order.SubOrderId)
			return err
		}
		return nil
	}

	// for cvm_separate_campus case
	if err := g.generateCVMSeparate(kt, order, existDevices); err != nil {
		logs.Errorf("failed to generate cvm in separate zones in region %s, suborder id: %s", order.Spec.Region,
			order.SubOrderId)
		return err
	}

	return nil
}

// retryMatchDevice retry to match generated devices
func (g *Generator) retryMatchDevice(devices []*types.DeviceInfo) error {
	genIDs := make([]int64, 0)
	for _, device := range devices {
		if !device.IsDelivered {
			genIDs = append(genIDs, int64(device.GenerateId))
		}
	}

	genIDs = utils.IntArrayUnique(genIDs)
	// update generate record to unmatched
	for _, genID := range genIDs {
		filter := &mapstr.MapStr{
			"generate_id": genID,
		}

		doc := mapstr.MapStr{
			"is_matched": false,
			"update_at":  time.Now(),
		}

		err := model.Operation().GenerateRecord().UpdateGenerateRecord(context.Background(), filter, &doc)
		if err != nil {
			logs.Errorf("failed to update generate record, generate id: %d, update: %+v, err: %v", genID, doc, err)
			return err
		}
	}

	return nil
}

// generateCVMConcentrate generates cvm devices in certain zone
func (g *Generator) generateCVMConcentrate(kt *kit.Kit, order *types.ApplyOrder,
	existDevices []*types.DeviceInfo) error {

	replicas := order.Total - uint(len(existDevices))
	// launch cvm
	if _, err := g.launchCvm(kt, order, order.Spec.Zone, replicas); err != nil {
		logs.Errorf("failed to launch cvm, err: %v", err)
		return err
	}
	return nil
}

// generateCVMSeparate generates cvm devices in separate zones
func (g *Generator) generateCVMSeparate(kt *kit.Kit, order *types.ApplyOrder, existDevices []*types.DeviceInfo) error {
	// 1. sum up each zone created devices
	createdTotalCount := uint(0)
	zoneCreatedCount := make(map[int]uint, 0)
	for _, device := range existDevices {
		zoneCreatedCount[device.ZoneID]++
		createdTotalCount++
	}

	// 2. get available zones
	requireType := order.RequireType
	// transform 4:"故障替换" to 1:"常规项目"
	if requireType == 4 {
		requireType = 1
	}
	availZones, err := g.getAvailableZoneInfo(kt, requireType, order.Spec.DeviceType, order.Spec.Region)
	if err != nil {
		logs.Errorf("failed to generate cvm, for get available zones err: %v, order id: %s", err, order.SubOrderId)
		return fmt.Errorf("failed to generate cvm, for get available zones err: %v", err)
	}
	if len(availZones) == 0 {
		logs.Errorf("failed to generate cvm, for get no available zones, order id: %s", order.SubOrderId)
		return fmt.Errorf("failed to generate cvm, for get no available zones")
	}

	// 3. get capacity
	zoneCapacity, err := g.getCapacity(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region,
		cvmapi.CvmSeparateCampus, "", "")
	if err != nil {
		logs.Errorf("failed to generate cvm, for get zone capacity err: %v, order id: %s", err, order.SubOrderId)
		return fmt.Errorf("failed to generate cvm, for get zone capacity err: %v", err)
	}
	logs.Infof("zone capacity: %+v, order id: %s", zoneCapacity, order.SubOrderId)

	// 4. for each zone, calculate replicas and launch cvm
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	genRecordIds := make([]uint64, 0)
	appendError := func(err error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, err)
	}
	appendGenRecord := func(id uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		genRecordIds = append(genRecordIds, id)
	}
	maxCount := math.Ceil(float64(order.Total) / 2)
	for _, zone := range availZones {
		replicas := uint(0)
		if len(availZones) > 1 {
			// 一个城市有大于一个campus的话，该campus最多只能生产需求数量的一半
			// 若单据无法完成，则剩余不生产，等人工介入处理
			campusMax := math.Max(
				maxCount-float64(zoneCreatedCount[int(zone.CmdbZoneId)]),
				0)
			replicas = uint(math.Min(
				math.Min(float64(order.Total-createdTotalCount), float64(zoneCapacity[zone.Zone])),
				campusMax))
		} else {
			// 一个城市只有一个campus的话，全部生产
			replicas = uint(math.Min(
				math.Min(float64(order.Total-createdTotalCount), float64(zoneCapacity[zone.Zone])),
				maxCount))
		}

		if replicas <= 0 {
			continue
		}

		zoneCreatedCount[int(zone.CmdbZoneId)] += replicas
		createdTotalCount += replicas

		wg.Add(1)
		go func(order *types.ApplyOrder, zoneId string, replicas uint) {
			defer wg.Done()
			genId, err := g.launchCvm(kt, order, zoneId, replicas)
			if err != nil {
				logs.Errorf("failed to launch cvm, err: %v", err)
				appendError(err)
			} else {
				logs.Infof("success to launch cvm, zone: %s, replicas: %d, order id: %s, generate id: %d", zoneId,
					replicas, order.SubOrderId, genId)
				appendGenRecord(genId)
			}
		}(order, zone.Zone, replicas)

		if order.Total <= createdTotalCount {
			break
		}
	}

	wg.Wait()

	if len(genRecordIds) == 0 {
		logs.Errorf("failed to generate cvm separate, for no zone has generate record")
		return fmt.Errorf("failed to generate cvm separate, for no zone has generate record")
	}

	if len(errs) > 0 {
		logs.Warnf("failed to generate cvm separate, errs: %v", errs)

		// check all generate records and update apply order status
		if err := g.checkGenerateRecordByOrder(order.SubOrderId); err != nil {
			logs.Warnf("failed to check generate record by order %s, err: %v", order.SubOrderId, err)
		}
	}

	return nil
}

func (g *Generator) checkGenerateRecordByOrder(suborderID string) error {
	genRecords, err := g.getOrderGenRecords(suborderID)
	if err != nil {
		logs.Errorf("failed to get generate records, order id: %s, err: %v", suborderID, err)
		return err
	}

	hasGenRecordMatching := false
	for _, record := range genRecords {
		if record.Status == types.GenerateStatusHandling ||
			record.Status == types.GenerateStatusSuccess && record.IsMatched == false {
			hasGenRecordMatching = true
			break
		}
	}

	stage := types.TicketStageRunning
	status := types.ApplyStatusMatchedSome
	if hasGenRecordMatching {
		status = types.ApplyStatusMatching
	}

	// do update apply order status
	filter := &mapstr.MapStr{
		"suborder_id": suborderID,
	}

	doc := &mapstr.MapStr{
		"stage":     stage,
		"status":    status,
		"update_at": time.Now(),
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to update apply order, id: %s, err: %v", suborderID, err)
		return err
	}

	return nil
}

// GenerateDVM generates docker vm devices
func (g *Generator) GenerateDVM(kt *kit.Kit, order *types.ApplyOrder) error {
	// 1. 解析请求结构
	selector, err := g.parseDvmSelector(kt, order)
	if err != nil {
		logs.Errorf("failed to parse dvm selector, err: %v, order id: %s", err, order.SubOrderId)
		return err
	}

	// 2. get history generated devices
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	existCount := uint(len(existDevices))
	if existCount >= order.Total {
		logs.Infof("apply order %s has been scheduled %d docker vm", order.SubOrderId, existCount)
		return nil
	}
	logs.Infof("apply order %s existing device number: %d", order.SubOrderId, existCount)

	// 3. 初始化（存量设备）亲和性
	// 记录每类亲和维度的设备数
	// 如果亲和性有要求时，每个维度(campus\module...)的设备不能超过一半
	antiAffinityReplicas := make(map[string]uint)
	for _, device := range existDevices {
		antiAffinityReplicas[g.antiAffinityValue(order.AntiAffinityLevel, types.HostPriority{
			IP:         device.Ip,
			SZone:      device.ZoneName,
			Equipment:  device.Equipment,
			ModuleName: strings.ToLower(device.ModuleName),
		})]++
	}

	// 4. 计算(一个子单)虚拟比不超过1:3
	existingHostMap := make(map[string]*dvmapi.DockerHost)
	for _, host := range existDevices {
		hostAssetID := ""
		parts := strings.Split(host.AssetId, "-")
		if len(parts) > 1 {
			hostAssetID = parts[0]
		}
		if val, ok := existingHostMap[hostAssetID]; ok {
			val.ScheduledVMs++
			existingHostMap[hostAssetID] = val
		} else {
			existingHostMap[hostAssetID] = &dvmapi.DockerHost{
				ScheduledVMs: 1,
				AssetID:      hostAssetID,
			}
		}
	}

	// 5. get allocatable docker hosts
	allocatableHosts, err := g.getAllocatableHosts(selector, order.ResourceType, existingHostMap)
	if err != nil {
		logs.Errorf("failed to get allocatable hosts, err: %v, order id: %s", err, order.SubOrderId)
		return err
	}
	if len(allocatableHosts) == 0 {
		logs.Errorf("get no allocatable hosts, order id: %s", order.SubOrderId)
		return fmt.Errorf("get no allocatable hosts")
	}

	// 6. sort allocatable docker hosts
	sortList := g.sortHosts(order.AntiAffinityLevel, allocatableHosts)
	logs.V(4).Infof("allocatable host list: %+v", sortList)

	// 7. try launch docker vm in order
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	genRecordIds := make([]uint64, 0)
	appendError := func(err error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, err)
	}
	appendGenRecord := func(id uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		genRecordIds = append(genRecordIds, id)
	}

	maxCount := order.Total
	if order.AntiAffinityLevel != types.AntiNone {
		maxCount = uint(math.Ceil(float64(order.Total) / 2))
		if maxCount == 0 {
			maxCount = 1
		}
	}
	for _, host := range sortList {
		// 计算最多可生产的容器数
		existNum := g.sumReplicas(antiAffinityReplicas)
		replicas := uint(math.Min(
			math.Min(
				math.Min(
					// 还需要生产的数量
					float64(order.Total-existNum),
					// 每台母机剩余的可生产数
					float64(host.AllocatableCount)),
				// 最大虚拟比
				float64(maxVirtualRatio-host.ScheduledVMs)),
			// 亲和性最大可生产数
			float64(maxCount-antiAffinityReplicas[g.antiAffinityValue(order.AntiAffinityLevel, host)]),
		))

		logs.V(5).Infof("host %s, module name: %s, total: %d, created: %d, allocatable: %d, observed replicas: %d",
			host.IP, host.ModuleName, order.Total, existNum, host.AllocatableCount, replicas)

		if replicas <= 0 {
			continue
		}

		antiAffinityReplicas[g.antiAffinityValue(order.AntiAffinityLevel, host)] += replicas

		// launch docker vm
		wg.Add(1)
		go func(order *types.ApplyOrder, selector *types.DVMSelector, host *types.HostPriority, replicas uint) {
			defer wg.Done()
			genId, err := g.launchDvm(order, selector, host, replicas)
			if err != nil {
				logs.Errorf("failed to launch dvm, err: %v", err)
				appendError(err)
			} else {
				logs.Infof("success to launch dvm, host: %+v, replicas: %d, order id: %s, generate id: %d", host,
					replicas, order.SubOrderId, genId)
				appendGenRecord(genId)
			}
		}(order, selector, &host, replicas)

		if order.Total <= g.sumReplicas(antiAffinityReplicas) {
			break
		}
	}

	wg.Wait()

	if len(errs) > 0 {
		logs.Warnf("failed to generate dvm, errs: %v", errs)
	}

	if len(genRecordIds) == 0 {
		logs.Errorf("failed to generate dvm, for no host has generate record")
		return fmt.Errorf("failed to generate dvm, for no host has generate record")
	}

	return nil
}

func (g *Generator) parseDvmSelector(kt *kit.Kit, order *types.ApplyOrder) (*types.DVMSelector, error) {
	req := &cfgtypes.GetDeviceParam{
		Filter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "device_type",
						Operator: querybuilder.OperatorEqual,
						Value:    order.Spec.DeviceType,
					}},
			},
		},
		Page: metadata.BasePage{
			Limit: 1,
			Start: 0,
		},
	}

	rst, err := g.configLogics.Device().GetDvmDeviceType(kt, req)
	if err != nil {
		logs.Errorf("failed to get dvm device info, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("failed to get device info, err: %v", err)
	}
	cnt := len(rst.Info)
	if cnt != 1 {
		logs.Errorf("failed to get dvm device info, for invalid info cnt %d != 1, rid: %s", cnt, kt.Rid)
		return nil, fmt.Errorf("failed to get dvm device info, for invalid info cnt %d != 1", cnt)
	}

	deviceGroup, err := rst.Info[0].Label.String("device_group")
	if err != nil {
		logs.Errorf("failed to get dvm device info, for invalid label.device_group %v is not string, err: %v, rid: %s",
			rst.Info[0].Label, err, kt.Rid)
		return nil, fmt.Errorf("failed to get dvm device info, for invalid label.device_group %v is not string",
			rst.Info[0].Label)
	}
	selector := &types.DVMSelector{
		Cores:             int(rst.Info[0].Cpu),
		Memory:            int(rst.Info[0].Mem),
		Disk:              int(rst.Info[0].Disk),
		DeviceClass:       rst.Info[0].DeviceType,
		Image:             order.Spec.Image,
		Kernel:            order.Spec.Kernel,
		DockerType:        deviceGroup,
		NetworkType:       rst.Info[0].NetWork,
		DataDiskMountPath: order.Spec.MountPath,
		DataDiskType:      order.Spec.DiskType,
		DataDiskRaid:      order.Spec.RaidType,
		Region:            order.Spec.Region,
		Zone:              order.Spec.Zone,
		ExtranetIsp:       order.Spec.Isp,
		CpuProvider:       rst.Info[0].CpuProvider,
	}

	// set network type TENTHOUSAND by default
	if selector.NetworkType == "" {
		selector.NetworkType = "TENTHOUSAND"
	}

	// get amd device pattern
	if selector.CpuProvider != "" {
		selector.AmdDevicePattern = g.getAMDDevicePattern()
	}
	selector.HostRole = g.getSpecialAppRole(strconv.Itoa(int(order.BkBizId)))

	return selector, nil
}

// getUnreleasedDevice gets unreleased devices bindings to current apply order
func (g *Generator) getUnreleasedDevice(orderId string) ([]*types.DeviceInfo, error) {
	filter := &mapstr.MapStr{
		"suborder_id": orderId,
		//"is_delivered": true,
	}

	devices, err := model.Operation().DeviceInfo().GetDeviceInfo(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get binding devices to order %s, err: %v", orderId, err)
	}

	return devices, nil
}

// launchCvm creates cvm and return created device ips
func (g *Generator) launchCvm(kt *kit.Kit, order *types.ApplyOrder, zone string, replicas uint) (uint64, error) {
	// 1. init generate record
	generateId, err := g.initGenerateRecord(order.ResourceType, order.SubOrderId, replicas)
	if err != nil {
		logs.Errorf("failed to launch cvm when init generate record, order id: %s, err: %v", order.SubOrderId,
			err)
		return 0, fmt.Errorf("failed to launch cvm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 2. launch cvm request
	request, err := g.buildCvmReq(kt, order, zone, replicas)
	if err != nil {
		logs.Errorf("failed to launch cvm when build cvm request, err: %v, order id: %s", err, order.SubOrderId)

		// update generate record status to Done
		if errRecord := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusFailed, err.Error(),
			"", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, err: %v", order.SubOrderId,
				errRecord)
			return generateId, fmt.Errorf("failed to launch cvm, order id: %s, err: %v", order.SubOrderId, errRecord)
		}

		return generateId, fmt.Errorf("failed to launch cvm, order id: %s, err: %v", order.SubOrderId, err)
	}

	taskId, err := g.createCVM(request)
	if err != nil {
		logs.Errorf("failed to launch cvm when create generate task, order id: %s, err: %v", order.SubOrderId,
			err)

		// update generate record status to Done
		if errRecord := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusFailed, err.Error(),
			"", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to launch cvm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 3. update generate record status to Query
	if err := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusHandling, "handling", taskId,
		nil); err != nil {
		logs.Errorf("failed to launch cvm when update generate record, order id: %s, err: %v", order.SubOrderId,
			err)
		return generateId, fmt.Errorf("failed to launch cvm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 4. check cvm task result
	if err = g.checkCVM(taskId); err != nil {
		logs.Errorf("failed to create cvm when check generate task, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)

		// update generate record status to Done
		if errRecord := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusFailed, err.Error(),
			"", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
			taskId, err)
	}

	// 5. get generated cvm instances
	hosts, err := g.listCVM(taskId)
	if err != nil {
		logs.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v", order.SubOrderId, taskId, err)

		// update generate record status to Done
		if errRecord := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusFailed, err.Error(),
			"", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
	}

	deviceList := make([]*types.DeviceInfo, 0)
	successIps := make([]string, 0)
	for _, host := range hosts {
		deviceList = append(deviceList, &types.DeviceInfo{
			Ip:               host.LanIp,
			AssetId:          host.AssetId,
			GenerateTaskId:   taskId,
			GenerateTaskLink: cvmapi.CvmOrderLinkPrefix + taskId,
			Deliverer:        "icr",
		})
		successIps = append(successIps, host.LanIp)
	}

	// 6. save generated cvm instances info
	if err := g.updateGeneratedDevice(order, generateId, deviceList); err != nil {
		logs.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
		return generateId, fmt.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 7. update generate record status to success
	if err := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusSuccess, "success", "",
		successIps); err != nil {
		logs.Errorf("failed to launch cvm when update generate record, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
		return generateId, fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
			taskId, err)
	}

	return generateId, nil
}

// launchDvm creates docker vm and return created device ips
func (g *Generator) launchDvm(order *types.ApplyOrder, applyRequest *types.DVMSelector, host *types.HostPriority,
	replicas uint) (uint64, error) {

	// 1. init generate record
	generateId, err := g.initGenerateRecord(order.ResourceType, order.SubOrderId, replicas)
	if err != nil {
		logs.Errorf("failed to launch docker vm when init generate record, order id: %s, err: %v", order.SubOrderId,
			err)
		return 0, fmt.Errorf("failed to launch docker vm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 2. launch dvm request
	request := &dvmapi.OrderCreateReq{
		Cores:     applyRequest.Cores,
		Memory:    applyRequest.Memory,
		Disk:      applyRequest.Disk,
		Image:     applyRequest.Image,
		SetId:     host.SetId,
		HostType:  host.DeviceClass,
		HostIp:    []string{host.IP},
		Affinity:  0,
		MountPath: applyRequest.DataDiskMountPath,
		Replicas:  replicas,
		Operator:  order.User,
		Module:    applyRequest.Zone,
		// 资源运营服务
		DisplayName: "931",
		// 开发测试-SCR_加工池
		AppModuleName: "51524",
		Reason:        order.Remark,
		HostRole:      applyRequest.HostRole,
	}

	taskId, err := g.createDVM(request)
	if err != nil {
		logs.Errorf("failed to create docker vm when launch generate task, order id: %s, err: %v", order.SubOrderId,
			err)

		// update generate record status to Done
		if errRecord := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusFailed, err.Error(),
			"", nil); errRecord != nil {
			logs.Errorf("failed to create dvm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch dvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to create docker vm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 3. update generate record status to Query
	if err := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusHandling, "handling", taskId,
		nil); err != nil {
		logs.Errorf("failed to launch docker vm when update generate record, order id: %s, err: %v", order.SubOrderId,
			err)
		return generateId, fmt.Errorf("failed to launch docker vm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 4. check cvm task result
	if err = g.checkDVM(taskId); err != nil {
		logs.Errorf("failed to launch docker vm when check generate task, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)

		// update generate record status to Done
		if errRecord := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusFailed, err.Error(),
			"", nil); errRecord != nil {
			logs.Errorf("failed to launch docker vm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch docker vm, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to launch docker vm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
	}

	// 5. get generated cvm instances
	hosts, err := g.listDVM(taskId)
	if err != nil {
		logs.Errorf("failed to list created docker vm, order id: %s, task id: %s, err: %v", order.SubOrderId, taskId,
			err)

		// update generate record status to Done
		if errRecord := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusFailed, err.Error(),
			"", nil); errRecord != nil {
			logs.Errorf("failed to create dvm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch dvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to list created docker vm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
	}

	deviceList := make([]*types.DeviceInfo, 0)
	successIps := make([]string, 0)
	for _, host := range hosts {
		if len(host.IP) <= 0 {
			continue
		}
		deviceList = append(deviceList, &types.DeviceInfo{
			Ip:               host.IP,
			GenerateTaskId:   taskId,
			GenerateTaskLink: fmt.Sprintf(dvmapi.DvmOrderLinkFormat, taskId),
			Deliverer:        "icr",
		})
		successIps = append(successIps, host.IP)
	}

	// 6. save generated cvm instances info
	if err := g.updateGeneratedDevice(order, generateId, deviceList); err != nil {
		logs.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
		return generateId, fmt.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 7. update generate record status to WaitForMatch
	if err := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusSuccess, "success", "",
		successIps); err != nil {
		logs.Errorf("failed to launch docker vm when update generate record, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
		return generateId, fmt.Errorf("failed to launch docker vm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
	}

	return generateId, nil
}

// getOrderGenRecords gets all generate records related to given order
func (g *Generator) getOrderGenRecords(suborderID string) ([]*types.GenerateRecord, error) {
	filter := map[string]interface{}{
		"suborder_id": suborderID,
	}
	page := metadata.BasePage{
		Start: 0,
		Limit: common.BKNoLimit,
	}

	records, err := model.Operation().GenerateRecord().FindManyGenerateRecord(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get generate record by order id: %s", suborderID)
		return nil, err
	}

	return records, nil
}

// initGenerateRecord creates generate record
func (g *Generator) initGenerateRecord(resourceType types.ResourceType, orderId string, total uint) (uint64, error) {
	id, err := model.Operation().GenerateRecord().NextSequence(context.Background())
	if err != nil {
		logs.Errorf("failed to get generate record next sequence id, order id: %s, err: %v", err)
		return 0, err
	}

	now := time.Now()
	record := &types.GenerateRecord{
		SubOrderId:   orderId,
		GenerateId:   id,
		GenerateType: string(resourceType),
		Status:       types.GenerateStatusInit,
		IsMatched:    false,
		TotalNum:     total,
		SuccessNum:   0,
		SuccessList:  make([]string, 0),
		CreateAt:     now,
		UpdateAt:     now,
		StartAt:      now,
	}

	if err := model.Operation().GenerateRecord().CreateGenerateRecord(context.Background(), record); err != nil {
		logs.Errorf("failed to init generate record, order id: %s, err: %v", orderId, err)
		return 0, err
	}

	return id, nil
}

// updateGenerateRecord updates generate record
func (g *Generator) updateGenerateRecord(resourceType types.ResourceType, generateId uint64,
	status types.GenerateStepStatus, msg, vmTaskId string, ipList []string) error {

	// TODO: filter add last status
	filter := &mapstr.MapStr{
		"generate_id": generateId,
	}

	now := time.Now()
	doc := mapstr.MapStr{
		"status":    status,
		"update_at": now,
	}

	if len(msg) != 0 {
		doc["message"] = msg
	}

	if len(vmTaskId) != 0 {
		link := ""
		switch resourceType {
		case types.ResourceTypePool:
			link = PoolOrderLinkPrefix + vmTaskId
		case types.ResourceTypeCvm:
			link = cvmapi.CvmOrderLinkPrefix + vmTaskId
		case types.ResourceTypeQcloudDvm, types.ResourceTypeIdcDvm:
			link = fmt.Sprintf(dvmapi.DvmOrderLinkFormat, vmTaskId)
		}
		doc["task_id"] = vmTaskId
		doc["task_link"] = link
	}

	if ipList != nil && len(ipList) > 0 {
		doc["success_num"] = len(ipList)
		doc["success_list"] = ipList
	}

	if status == types.GenerateStatusFailed || status == types.GenerateStatusSuccess {
		doc["end_at"] = now
	}

	if err := model.Operation().GenerateRecord().UpdateGenerateRecord(context.Background(), filter, &doc); err != nil {
		logs.Errorf("failed to update generate record, generate id: %d, update: %+v, err: %v", generateId, doc, err)
		return err
	}

	return nil
}

func (g *Generator) updateGeneratedDevice(order *types.ApplyOrder, generateId uint64, items []*types.DeviceInfo) error {
	ips := make([]string, 0)
	assetIds := make([]string, 0)
	for _, item := range items {
		ips = append(ips, item.Ip)
		assetIds = append(assetIds, item.AssetId)
	}

	// 1. sync device info to cc
	if order.ResourceType == types.ResourceTypeCvm {
		if err := g.syncHostByAsset(assetIds); err != nil {
			logs.Errorf("failed to sync device info to cc, order id: %s, err: %v", order.SubOrderId, err)
			return err
		}
	} else {
		if err := g.syncHostByIp(ips); err != nil {
			logs.Errorf("failed to sync device info to cc, order id: %s, err: %v", order.SubOrderId, err)
			return err
		}
	}

	// 2. get cc host detail info
	ccHosts, err := g.getHostDetail(ips)
	if err != nil {
		logs.Errorf("failed to get cc host info, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}
	mapIpToHost := make(map[string]*cmdb.HostInfo)
	for _, host := range ccHosts {
		mapIpToHost[host.GetUniqIp()] = host
	}

	// 3. save device info to db
	now := time.Now()
	for _, item := range items {
		if isDup, _ := g.isDuplicateHost(order.SubOrderId, item.Ip); isDup {
			logs.Warnf("duplicate host for order id: %s, ip: %s", order.SubOrderId, item.Ip)
			continue
		}
		device := &types.DeviceInfo{
			OrderId:      order.OrderId,
			SubOrderId:   order.SubOrderId,
			GenerateId:   generateId,
			BkBizId:      int(order.BkBizId),
			User:         order.User,
			Ip:           item.Ip,
			RequireType:  order.RequireType,
			ResourceType: order.ResourceType,
			// set device type according to order specification by default
			DeviceType:        order.Spec.DeviceType,
			Description:       order.Description,
			Remark:            order.Remark,
			IsMatched:         false,
			IsChecked:         false,
			IsInited:          false,
			IsDiskChecked:     false,
			IsDelivered:       false,
			GenerateTaskId:    item.GenerateTaskId,
			GenerateTaskLink:  item.GenerateTaskLink,
			InitTaskId:        item.InitTaskId,
			InitTaskLink:      item.InitTaskLink,
			DiskCheckTaskId:   item.DiskCheckTaskId,
			DiskCheckTaskLink: item.DiskCheckTaskLink,
			Deliverer:         item.Deliverer,
			CreateAt:          now,
			UpdateAt:          now,
		}
		// add device detail info from cc
		if host, ok := mapIpToHost[item.Ip]; !ok {
			logs.Warnf("failed to get %s detail info in cc", item.Ip)
		} else {
			device.AssetId = host.BkAssetId
			// update device type from cc
			device.DeviceType = host.SvrDeviceClass
			device.ZoneName = host.SubZone
			zoneId, err := strconv.Atoi(host.SubZoneId)
			if err != nil {
				logs.Warnf("failed to convert sub zone id %s to int", host.SubZoneId)
				device.ZoneID = 0
			} else {
				device.ZoneID = zoneId
			}
			device.ModuleName = host.ModuleName
			device.Equipment = host.RackId
		}

		if err := model.Operation().DeviceInfo().CreateDeviceInfo(context.Background(), device); err != nil {
			logs.Errorf("failed to save device info to db, order id: %s, err: %v", order.SubOrderId, err)
			return err
		}
	}

	logs.Infof("successfully sync device info to cc, ips: %+v, assets: %+v", ips, assetIds)

	return nil
}

func (g *Generator) isDuplicateHost(suborderId, ip string) (bool, error) {
	filter := map[string]interface{}{
		"suborder_id": suborderId,
		"ip":          ip,
	}

	cnt, err := model.Operation().DeviceInfo().CountDeviceInfo(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to count device info, order id: %s, ip: %s, err: %v", suborderId, ip, err)
		return false, err
	}

	if cnt >= 1 {
		return true, nil
	}

	return false, nil
}

func (g *Generator) getHostDetail(ips []string) ([]*cmdb.HostInfo, error) {
	req := &cmdb.ListBizHostReq{
		BkBizId: 931,
		BkModuleIds: []int64{
			// RA池
			239148,
			// SA云化池
			239149,
			// SCR_加工池
			532040,
		},
		HostPropertyFilter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_host_innerip",
						Operator: querybuilder.OperatorIn,
						Value:    ips,
					},
					// support bk_cloud_id 0 only
					querybuilder.AtomRule{
						Field:    "bk_cloud_id",
						Operator: querybuilder.OperatorEqual,
						Value:    0,
					},
				},
			},
		},
		Fields: []string{
			"bk_host_id",
			"bk_asset_id",
			"bk_host_innerip",
			"bk_host_outerip",
			// 外网运营商
			"bk_ip_oper_name",
			// 机型
			"svr_device_class",
			"bk_os_name",
			// 地域
			"bk_zone_name",
			// 可用区(子Zone)
			"sub_zone",
			// 子ZoneID
			"sub_zone_id",
			"module_name",
			// 机架号，string类型
			"rack_id",
			"idc_unit_name",
			// 逻辑区域
			"logic_domain",
			"raid_name",
			"svr_input_time",
		},
		Page: cmdb.BasePage{
			Start: 0,
			Limit: common.BKMaxInstanceLimit,
		},
	}

	resp, err := g.cc.ListBizHost(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	if resp.Result == false || resp.Code != 0 {
		logs.Errorf("failed to get cc host info, code: %d, msg: %s", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("failed to get cc host info, err: %s", resp.ErrMsg)
	}

	return resp.Data.Info, nil
}

// MatchCVM manual match cvm devices
func (g *Generator) MatchCVM(param *types.MatchDeviceReq) error {
	// 1. get order by suborder id
	order, err := g.getApplyOrder(param.SuborderId)
	if err != nil {
		logs.Errorf("failed to match cvm when get apply order, err: %v, order id: %s", err, param.SuborderId)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, param.SuborderId)
	}

	// cannot match device if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot match device, for order %s stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
		return fmt.Errorf("cannot match device, for order %s stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
	}

	// set apply order status MATCHING
	if err := g.lockApplyOrder(order); err != nil {
		logs.Errorf("failed to match cvm when lock apply order, err: %v, order id: %s", err, param.SuborderId)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, param.SuborderId)
	}

	replicas := uint(len(param.Device))

	// 2. init generate record
	generateId, err := g.initGenerateRecord(order.ResourceType, order.SubOrderId, replicas)
	if err != nil {
		logs.Errorf("failed to match cvm when init generate record, err: %v, order id: %s", err, order.SubOrderId)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, order.SubOrderId)
	}

	// TODO: check whether device is locked by other orders
	deviceList := make([]*types.DeviceInfo, 0)
	successIps := make([]string, 0)
	for _, host := range param.Device {
		deviceList = append(deviceList, &types.DeviceInfo{
			Ip:        host.Ip,
			AssetId:   host.AssetId,
			Deliverer: param.Operator,
		})
		successIps = append(successIps, host.Ip)
	}

	// 3. save generated cvm instances info
	if err := g.updateGeneratedDevice(order, generateId, deviceList); err != nil {
		logs.Errorf("failed to update generated device, err: %v, order id: %s", err, order.SubOrderId)
		return fmt.Errorf("failed to update generated device, err: %v, order id: %s", err, order.SubOrderId)
	}

	// 4. update generate record status to success
	msg := fmt.Sprintf("manually matched by %s successfully", param.Operator)
	if err := g.updateGenerateRecord(order.ResourceType, generateId, types.GenerateStatusSuccess, msg, "",
		successIps); err != nil {
		logs.Errorf("failed to match cvm when update generate record, err: %v, order id: %s", err, order.SubOrderId)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, order.SubOrderId)
	}

	return nil
}

// getApplyOrder gets apply order by order id
func (g *Generator) getApplyOrder(key string) (*types.ApplyOrder, error) {
	filter := &mapstr.MapStr{
		"suborder_id": key,
	}
	order, err := model.Operation().ApplyOrder().GetApplyOrder(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get apply order by id: %s", key)
		return nil, err
	}

	return order, nil
}

// lockApplyOrder locks apply order to avoid order repeat dispatch
func (g *Generator) lockApplyOrder(order *types.ApplyOrder) error {
	filter := &mapstr.MapStr{
		"suborder_id": order.SubOrderId,
	}

	doc := &mapstr.MapStr{
		"stage":     types.TicketStageRunning,
		"status":    types.ApplyStatusMatching,
		"update_at": time.Now(),
	}

	if err := model.Operation().ApplyOrder().UpdateApplyOrder(context.Background(), filter, doc); err != nil {
		logs.Errorf("failed to lock apply order, id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	return nil
}

// MatchPM automatically match physical machine devices
func (g *Generator) MatchPM(order *types.ApplyOrder) error {
	// 1. get history generated devices
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v", order.SubOrderId, err)
		return err
	}

	// 2. check if need generate device
	existCount := uint(len(existDevices))
	if existCount >= order.Total {
		logs.Infof("apply order %s has been scheduled %d pm", order.SubOrderId, existCount)
		return nil
	}

	logs.Infof("apply order %s existing device number: %d", order.SubOrderId, existCount)

	// 3. match pm
	if err := g.matchPM(order, existDevices); err != nil {
		logs.Errorf("failed to match pm, suborder id: %s", order.SubOrderId)
		return err
	}

	return nil
}

// MatchPoolDevice manual match pool devices
func (g *Generator) MatchPoolDevice(param *types.MatchPoolDeviceReq) error {
	// 1. get order by suborder id
	order, err := g.getApplyOrder(param.SuborderId)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v, order id: %s", err, param.SuborderId)
		return fmt.Errorf("failed to get apply order, err: %v, order id: %s", err, param.SuborderId)
	}

	// cannot match device if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot match device, for order %s stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
		return fmt.Errorf("cannot match device, for order %s stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
	}

	// set apply order status MATCHING
	if err := g.lockApplyOrder(order); err != nil {
		logs.Errorf("failed to lock apply order, err: %v, order id: %s", err, param.SuborderId)
		return fmt.Errorf("failed to lock apply order, err: %v, order id: %s", err, param.SuborderId)
	}

	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	genRecordIds := make([]uint64, 0)
	appendError := func(err error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, err)
	}
	appendGenRecord := func(id uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		genRecordIds = append(genRecordIds, id)
	}

	for _, task := range param.Spec {
		wg.Add(1)
		go func(order *types.ApplyOrder, recall *types.MatchPoolSpec) {
			defer wg.Done()
			newKit := kit.New()
			genId, err := g.launchRecallHost(newKit, order, recall)
			if err != nil {
				logs.Errorf("failed to launch recall order, err: %v", err)
				appendError(err)
			} else {
				logs.Infof("success to launch recall order, replicas: %d, order id: %s, generate id: %d",
					recall.Replicas, order.SubOrderId, genId)
				appendGenRecord(genId)
			}
		}(order, task)
	}

	wg.Wait()

	if len(errs) > 0 {
		logs.Warnf("failed to generate pool recall device, errs: %v", errs)
	}

	if len(genRecordIds) == 0 {
		logs.Errorf("failed to generate pool recall device, for no generate record")
		return fmt.Errorf("failed to generate pool recall device, for no generate record")
	}

	return nil
}
