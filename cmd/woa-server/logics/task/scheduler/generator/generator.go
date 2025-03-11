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

	"hcm/cmd/woa-server/logics/config"
	poolLogics "hcm/cmd/woa-server/logics/pool"
	rollingserver "hcm/cmd/woa-server/logics/rolling-server"
	"hcm/cmd/woa-server/logics/task/scheduler/algorithm"
	"hcm/cmd/woa-server/model/task"
	cfgtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/dal"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/thirdparty/dvmapi"
	"hcm/pkg/thirdparty/esb"
	"hcm/pkg/thirdparty/esb/cmdb"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
	utils "hcm/pkg/tools/util"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

// Generator generates vm devices
type Generator struct {
	cvm          cvmapi.CVMClientInterface
	dvm          dvmapi.DVMClientInterface
	cc           cmdb.Client
	ctx          context.Context
	configLogics config.Logics
	poolLogics   poolLogics.Logics
	rsLogics     rollingserver.Logics
	clientConf   cc.ClientConfig

	predicateFuncs map[string]algorithm.FitPredicate
	priorityFuncs  []algorithm.PriorityConfig
}

// New creates a generator
func New(ctx context.Context, rsLogics rollingserver.Logics, thirdCli *thirdparty.Client, esbCli esb.Client,
	clientConf cc.ClientConfig) (*Generator, error) {

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
		rsLogics:       rsLogics,
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
	genRecordIds, errs := g.batchLaunchCvm(kt, order, order.Spec.Zone, replicas)
	return g.checkLaunchCvmResult(kt, order.SubOrderId, genRecordIds, errs)
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
	availZones, err := g.getAvailableZoneInfo(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region)
	if err != nil {
		logs.Errorf("failed to generate cvm, for get available zones err: %v, order id: %s, rid: %s",
			err, order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to generate cvm, for get available zones err: %v", err)
	}
	if len(availZones) == 0 {
		logs.Errorf("failed to generate cvm, for get no available zones, order id: %s, rid: %s",
			order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to generate cvm, for get no available zones")
	}

	// 3. get capacity
	zoneCapacity, err := g.getCapacity(kt, order.RequireType, order.Spec.DeviceType, order.Spec.Region,
		cvmapi.CvmSeparateCampus, "", "", order.Spec.ChargeType)
	if err != nil {
		logs.Errorf("failed to generate cvm, for get zone capacity err: %v, order id: %s, rid: %s",
			err, order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to generate cvm, for get zone capacity err: %v", err)
	}

	logs.Infof("generateCVMSeparate campus start, subOrderID: %s, createdTotalCount: %d, zoneCapacity: %+v, "+
		"zoneCreatedCount: %v, availZones: %+v, rid: %s", order.SubOrderId, createdTotalCount, zoneCapacity,
		zoneCreatedCount, cvt.PtrToSlice(availZones), kt.Rid)

	// 4. for each zone, calculate replicas and launch cvm
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}
	errs := make([]error, 0)
	genRecordIds := make([]uint64, 0)
	appendError := func(subErrs []error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, subErrs...)
	}
	appendGenRecord := func(ids []uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		genRecordIds = append(genRecordIds, ids...)
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

		logs.Infof("generateCVMSeparate campus loop, subOrderID: %s, maxCount: %d, createdTotalCount: %d, "+
			"zoneCapacity: %+v, zoneCreatedCount: %v, zoneInfo: %+v, availZonesNum: %d, replicas: %d, rid: %s",
			order.SubOrderId, maxCount, createdTotalCount, zoneCapacity, zoneCreatedCount, cvt.PtrToVal(zone),
			len(availZones), replicas, kt.Rid)
		if replicas <= 0 {
			continue
		}

		zoneCreatedCount[int(zone.CmdbZoneId)] += replicas
		createdTotalCount += replicas

		wg.Add(1)
		go func(order *types.ApplyOrder, zoneId string, replicas uint) {
			defer wg.Done()
			genIds, subErrs := g.batchLaunchCvm(kt, order, zoneId, replicas)
			if len(subErrs) != 0 {
				logs.Errorf("failed to launch cvm, subOrderID: %s, subErrs: %v, zoneId: %s, rid: %s", order.SubOrderId,
					subErrs, zoneId, kt.Rid)
				appendError(subErrs)
			}
			if len(genIds) > 0 {
				logs.Infof("success to launch cvm, subOrderID: %s, zone: %s, generate ids: %v, rid: %s",
					order.SubOrderId, zoneId, genIds, kt.Rid)
				appendGenRecord(genIds)
			}
		}(order, zone.Zone, replicas)

		if order.Total <= createdTotalCount {
			break
		}
	}

	wg.Wait()

	return g.checkLaunchCvmResult(kt, order.SubOrderId, genRecordIds, errs)
}

func (g *Generator) checkLaunchCvmResult(kt *kit.Kit, subOrderID string, genRecordIds []uint64, errs []error) error {
	if len(genRecordIds) == 0 {
		logs.Errorf("failed to generate cvm, for no zone has generate record, subOrderID: %s, errs: %v, rid: %s",
			subOrderID, errs, kt.Rid)
		return fmt.Errorf("failed to generate cvm, for no zone has generate record")
	}

	if len(errs) > 0 {
		logs.Errorf("failed to generate cvm, subOrderID: %s, errs: %v, rid: %s", subOrderID, errs, kt.Rid)

		// check all generate records and update apply order status
		if err := g.UpdateOrderStatus(subOrderID); err != nil {
			logs.Errorf("failed to update order status, subOrderId: %s, err: %v, rid: %s", subOrderID, err, kt.Rid)
		}
	}

	return nil
}

// UpdateOrderStatus 更新订单状态
func (g *Generator) UpdateOrderStatus(suborderID string) error {
	genRecords, err := g.getOrderGenRecords(suborderID)
	if err != nil {
		logs.Errorf("failed to get generate records, order id: %s, err: %v", suborderID, err)
		return err
	}

	hasGenRecordMatching := false
	for _, record := range genRecords {
		if record.Status == types.GenerateStatusHandling ||
			record.Status == types.GenerateStatusSuccess && !record.IsMatched {
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
		// "is_delivered": true,
	}

	devices, err := model.Operation().DeviceInfo().GetDeviceInfo(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get binding devices to order %s, err: %v", orderId, err)
	}

	return devices, nil
}

// batchLaunchCvm  batch creates cvm and return created device ips
func (g *Generator) batchLaunchCvm(kt *kit.Kit, order *types.ApplyOrder, zone string, replicas uint) ([]uint64,
	[]error) {

	logs.Infof("start batch launch cvm, sub order id: %s, zone: %s, replicas: %d, rid: %s", order.SubOrderId, zone,
		replicas, kt.Rid)

	var requestNum uint
	excludeSubnetIDMap := make(map[string]struct{})
	generateIDs := make([]uint64, 0)
	errs := make([]error, 0)
	mutex := sync.Mutex{}
	appendGenRecord := func(id uint64) {
		mutex.Lock()
		defer mutex.Unlock()
		generateIDs = append(generateIDs, id)
	}
	appendError := func(err error) {
		mutex.Lock()
		defer mutex.Unlock()
		errs = append(errs, err)
	}
	eg := errgroup.Group{}
	eg.SetLimit(5)

	for replicas > requestNum {
		curRequiredNum := replicas - requestNum
		generateID, err := g.initGenerateRecord(order.ResourceType, order.SubOrderId, curRequiredNum, false)
		if err != nil {
			logs.Errorf("failed to launch cvm when init generate record, err: %v, sub order id: %s, rid: %s", err,
				order.SubOrderId, kt.Rid)
			appendError(fmt.Errorf("failed to launch cvm, sub order id: %s, err: %v", order.SubOrderId, err))
			break
		}

		createCvmReq, err := g.buildGenRecordCvmReq(kt, generateID, order, zone, curRequiredNum, excludeSubnetIDMap)
		if err != nil {
			logs.Errorf("failed to launch cvm when build cvm request, err: %v, generateID: %d, sub order id: %s, "+
				"rid: %s", err, generateID, order.SubOrderId, kt.Rid)
			appendError(fmt.Errorf("failed to launch cvm, sub order id: %s, err: %v", order.SubOrderId, err))
			break
		}
		excludeSubnetIDMap[createCvmReq.SubnetId] = struct{}{}
		requestNum += createCvmReq.ApplyNumber

		eg.Go(func() error {
			if err = g.launchCvm(kt, order, createCvmReq, generateID); err != nil {
				logs.Errorf("failed to launch cvm, err: %v, sub order id: %s, zone: %s, replicas: %d, generateID: %d,"+
					" rid: %s", err, order.SubOrderId, createCvmReq.Zone, createCvmReq.ApplyNumber, generateID, kt.Rid)
				appendError(err)
				return nil
			}
			logs.Infof("success to launch cvm, sub order id: %s, zone: %s, replicas: %d, generate id: %s, rid: %s",
				order.SubOrderId, createCvmReq.Zone, createCvmReq.ApplyNumber, generateID, kt.Rid)
			appendGenRecord(generateID)
			return nil
		})
	}

	_ = eg.Wait()

	return generateIDs, errs
}

func (g *Generator) buildGenRecordCvmReq(kt *kit.Kit, generateID uint64, order *types.ApplyOrder, zone string,
	replicas uint, excludeSubnetIDMap map[string]struct{}) (*types.CVM, error) {

	createCvmReq, err := g.buildCvmReq(kt, order, zone, replicas, excludeSubnetIDMap)
	if err != nil {
		logs.Errorf("failed to launch cvm when build cvm request, err: %v, order id: %s, rid: %s", err,
			order.SubOrderId, kt.Rid)
		// update generate record status to Done
		if subErr := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateID,
			types.GenerateStatusFailed, err.Error(), "", nil); subErr != nil {
			logs.Errorf("failed to create cvm when update generate record, err: %v, order id: %s, rid: %s",
				subErr, order.SubOrderId, kt.Rid)
			return nil, subErr
		}
		return nil, err
	}

	if createCvmReq.ApplyNumber == replicas {
		return createCvmReq, nil
	}

	filter := &mapstr.MapStr{
		"generate_id": generateID,
	}
	now := time.Now()
	doc := mapstr.MapStr{
		"total_num": createCvmReq.ApplyNumber,
		"update_at": now,
	}
	if err = model.Operation().GenerateRecord().UpdateGenerateRecord(kt.Ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update generate record, err: %v, generate id: %d, update: %+v, rid: %s", err, generateID,
			doc, kt.Rid)
		return nil, err
	}

	return createCvmReq, nil
}

// launchCvm creates cvm and return created device ips
func (g *Generator) launchCvm(kt *kit.Kit, order *types.ApplyOrder, createCvmReq *types.CVM, generateId uint64) error {
	taskId, err := g.createCVM(kt, createCvmReq, order)
	if err != nil {
		logs.Errorf("scheduler:logics:launch:cvm:failed, failed to launch cvm when create generate task, "+
			"order id: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskId, errRecord, kt.Rid)
			return fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return fmt.Errorf("failed to launch cvm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// update generate record status to Query
	if err = g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId, types.GenerateStatusHandling,
		"handling", taskId, nil); err != nil {
		logs.Errorf("scheduler:logics:launch:cvm:failed, failed to launch cvm when update generate record, "+
			"order id: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return fmt.Errorf("failed to launch cvm, order id: %s, err: %v", order.SubOrderId, err)
	}
	// check cvm task result and update generate record
	return g.AddCvmDevices(kt, taskId, generateId, order)
}

// AddCvmDevices check generated device, create device infos and update generate record status
func (g *Generator) AddCvmDevices(kt *kit.Kit, taskId string, generateId uint64,
	order *types.ApplyOrder) error {

	// 1. check cvm task result
	if err := g.checkCVM(taskId); err != nil {
		logs.Errorf("scheduler:logics:launch:cvm:failed, failed to create cvm when check generate task, "+
			"order id: %s, task id: %s, err: %v, rid: %s", order.SubOrderId, taskId, err, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(kt.Ctx, order.ResourceType, generateId, types.GenerateStatusFailed,
			err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskId, errRecord, kt.Rid)
			return fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
			taskId, err)
	}

	// 2. get generated cvm instances
	hosts, err := g.listCVM(taskId)
	if err != nil {
		logs.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v, rid: %s",
			order.SubOrderId, taskId, err, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(kt.Ctx, order.ResourceType, generateId, types.GenerateStatusFailed,
			err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskId, errRecord, kt.Rid)
			return fmt.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
		}

		return fmt.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)
	}
	// 3. create device infos
	return g.createDeviceInfo(kt, order, generateId, hosts, taskId)
}

func (g *Generator) createDeviceInfo(kt *kit.Kit, order *types.ApplyOrder, generateId uint64,
	hosts []*cvmapi.InstanceItem, taskId string) error {

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

	// NOTE: sleep 15 seconds to wait for CMDB host sync.
	time.Sleep(15 * time.Second)

	txnErr := dal.RunTransaction(kt, func(sc mongo.SessionContext) error {
		// 1. save generated cvm instances info
		sessionKit := &kit.Kit{Ctx: sc, Rid: kt.Rid}
		if err := g.createGeneratedDevices(sessionKit, order, generateId, deviceList); err != nil {
			logs.Errorf("failed to update generated device, order id: %s, err: %v, rid: %s", order.SubOrderId, err,
				kt.Rid)
			// update generate record status to Done
			// 不参与回滚
			if err := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
				types.GenerateStatusFailed, err.Error(), "", nil); err != nil {
				logs.Errorf("failed to update generate record, generate id: %d, err: %v, rid: %s", generateId, err,
					kt.Rid)
				return err
			}

			return fmt.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
		}

		// 2. update generate record status to success
		if err := g.UpdateGenerateRecord(sc, order.ResourceType, generateId, types.GenerateStatusSuccess, "success",
			"", successIps); err != nil {
			logs.Errorf("failed to launch cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskId, err, kt.Rid)
			return fmt.Errorf("failed to launch cvm, order id: %s, task id: %s, err: %v", order.SubOrderId, taskId, err)
		}

		return nil
	})

	if txnErr != nil {
		logs.Errorf("failed to launch cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
			order.SubOrderId, taskId, txnErr, kt.Rid)
		return fmt.Errorf("failed to launch cvm when update generate record, order id: %s, task id: %s, "+
			"err: %v", order.SubOrderId, taskId, txnErr)
	}
	return nil
}

// launchDvm creates docker vm and return created device ips
func (g *Generator) launchDvm(order *types.ApplyOrder, applyRequest *types.DVMSelector, host *types.HostPriority,
	replicas uint) (uint64, error) {

	// 1. init generate record
	generateId, err := g.initGenerateRecord(order.ResourceType, order.SubOrderId, replicas, false)
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
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to create dvm when update generate record, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskId, errRecord)
			return generateId, fmt.Errorf("failed to launch dvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskId, errRecord)
		}

		return generateId, fmt.Errorf("failed to create docker vm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 3. update generate record status to Query
	if err := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId, types.GenerateStatusHandling,
		"handling", taskId, nil); err != nil {
		logs.Errorf("failed to launch docker vm when update generate record, order id: %s, err: %v", order.SubOrderId,
			err)
		return generateId, fmt.Errorf("failed to launch docker vm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 4. check cvm task result
	if err = g.checkDVM(taskId); err != nil {
		logs.Errorf("failed to launch docker vm when check generate task, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskId, err)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
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
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
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
	if err := g.createGeneratedDevice(order, generateId, deviceList); err != nil {
		logs.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
		return generateId, fmt.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
	}

	// 7. update generate record status to WaitForMatch
	if err := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId, types.GenerateStatusSuccess,
		"success", "", successIps); err != nil {
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
		Limit: pkg.BKNoLimit,
	}

	records, err := model.Operation().GenerateRecord().FindManyGenerateRecord(context.Background(), page, filter)
	if err != nil {
		logs.Errorf("failed to get generate record by order id: %s", suborderID)
		return nil, err
	}

	return records, nil
}

// initGenerateRecord creates generate record
func (g *Generator) initGenerateRecord(resourceType types.ResourceType, subOrderId string, total uint,
	isManualMatched bool) (uint64, error) {

	id, err := model.Operation().GenerateRecord().NextSequence(context.Background())
	if err != nil {
		logs.Errorf("failed to get generate record next sequence id, subOrderId: %s, err: %v", subOrderId, err)
		return 0, err
	}

	now := time.Now()
	record := &types.GenerateRecord{
		SubOrderId:      subOrderId,
		GenerateId:      id,
		GenerateType:    string(resourceType),
		Status:          types.GenerateStatusInit,
		IsMatched:       false,
		TotalNum:        total,
		SuccessNum:      0,
		SuccessList:     make([]string, 0),
		CreateAt:        now,
		UpdateAt:        now,
		StartAt:         now,
		IsManualMatched: isManualMatched, // 是否手工匹配
	}

	if err = model.Operation().GenerateRecord().CreateGenerateRecord(context.Background(), record); err != nil {
		logs.Errorf("failed to init generate record, subOrderId: %s, err: %v", subOrderId, err)
		return 0, err
	}

	return id, nil
}

// UpdateGenerateRecord updates generate record
func (g *Generator) UpdateGenerateRecord(ctx context.Context, resourceType types.ResourceType,
	generateId uint64, status types.GenerateStepStatus, msg, vmTaskId string, ipList []string) error {

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

	if err := model.Operation().GenerateRecord().UpdateGenerateRecord(ctx, filter, &doc); err != nil {
		logs.Errorf("failed to update generate record, generate id: %d, update: %+v, err: %v", generateId, doc, err)
		return err
	}
	return nil
}

func (g *Generator) createGeneratedDevices(kt *kit.Kit, order *types.ApplyOrder, generateId uint64,
	items []*types.DeviceInfo) error {

	ips := make([]string, 0)
	assetIds := make([]string, 0)
	for _, item := range items {
		ips = append(ips, item.Ip)
		assetIds = append(assetIds, item.AssetId)
	}

	devices, err := g.syncHostToCMDB(order, generateId, items)
	if err != nil {
		logs.Errorf("failed to syn to cmdb, order id: %s, generateId: %d, err: %v, rid: %s", order.SubOrderId,
			generateId, err, kt.Rid)
		return err
	}

	if err = model.Operation().DeviceInfo().CreateDeviceInfos(kt.Ctx, devices); err != nil {
		logs.Errorf("failed to save device info to db, order id: %s, generateId: %d, err: %v, rid: %s",
			order.SubOrderId, generateId, err, kt.Rid)
		return err
	}

	logs.Infof("successfully sync device info to cc, orderId: %s, generateId: %d, ips: %+v, assets: %+v, "+
		"devices: %+v, rid: %s", order.SubOrderId, generateId, ips, assetIds, cvt.PtrToSlice(devices), kt.Rid)

	return nil
}

func (g *Generator) syncHostToCMDB(order *types.ApplyOrder, generateId uint64,
	items []*types.DeviceInfo) ([]*types.DeviceInfo, error) {
	ips := make([]string, 0)
	assetIds := make([]string, 0)
	for _, item := range items {
		ips = append(ips, item.Ip)
		assetIds = append(assetIds, item.AssetId)
	}

	// 线上Bug，返回了空的DeviceInfo数组，导致mongo插入失败
	if len(ips) == 0 && len(assetIds) == 0 {
		logs.Errorf("failed to sync device info to cc, ips and assetIds is empty, subOrderID: %s, generateId: %s, "+
			"items: %+v", order.SubOrderId, generateId, cvt.PtrToSlice(items))
		return nil, errf.Newf(errf.RecordNotFound, "failed to sync device info to cc, ips and assetIds is empty, "+
			"subOrderID: %s", order.SubOrderId)
	}

	// 1. sync device info to cc
	if order.ResourceType == types.ResourceTypeCvm {
		if err := g.syncHostByAsset(assetIds); err != nil {
			logs.Errorf("failed to sync device info to cc, order id: %s, err: %v, rid: %s", order.SubOrderId, err)
			return nil, err
		}
	} else {
		if err := g.syncHostByIp(ips); err != nil {
			logs.Errorf("failed to sync device info to cc, order id: %s, err: %v, rid: %s", order.SubOrderId, err)
			return nil, err
		}
	}

	// 2. get cc host detail info
	// 由于会存在主机在cc，但是此时机器还没有ip, 所以需要通过固资号进行查询
	ccHosts, err := g.getHostDetail(assetIds)
	if err != nil {
		logs.Errorf("failed to get cc host info, order id: %s, err: %v, rid: %s", order.SubOrderId, err)
		return nil, err
	}
	mapAssetIDToHost := make(map[string]*cmdb.Host)
	for _, host := range ccHosts {
		mapAssetIDToHost[host.BkAssetID] = host
	}

	logs.Infof("successfully sync device info to cc, subOrderID: %s, ips: %+v, assets: %+v",
		order.SubOrderId, ips, assetIds)
	devices := g.buildDevicesInfo(items, order, generateId, mapAssetIDToHost)
	return devices, nil
}

func (g *Generator) buildDevicesInfo(items []*types.DeviceInfo, order *types.ApplyOrder, generateId uint64,
	mapAssetIDToHost map[string]*cmdb.Host) []*types.DeviceInfo {

	// save device info to db
	now := time.Now()
	var devices []*types.DeviceInfo

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
			IsManualMatched:   item.IsManualMatched,
			CreateAt:          now,
			UpdateAt:          now,
		}
		// add device detail info from cc
		if host, ok := mapAssetIDToHost[item.AssetId]; !ok {
			logs.Warnf("failed to get %s detail info in cc", item.AssetId)
		} else {
			device.AssetId = host.BkAssetID
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

		devices = append(devices, device)
	}
	return devices
}

func (g *Generator) createGeneratedDevice(order *types.ApplyOrder, generateId uint64, items []*types.DeviceInfo) error {
	ips := make([]string, 0)
	assetIds := make([]string, 0)
	for _, item := range items {
		ips = append(ips, item.Ip)
		assetIds = append(assetIds, item.AssetId)
	}

	devices, err := g.syncHostToCMDB(order, generateId, items)
	if err != nil {
		logs.Errorf("failed to syn to cmdb, order id: %s, generateId: %d, err: %v", order.SubOrderId, generateId, err)
		return err
	}

	if err = model.Operation().DeviceInfo().CreateDeviceInfos(context.Background(), devices); err != nil {
		logs.Errorf("failed to save device info to db, order id: %s, generateId: %d, err: %v", order.SubOrderId,
			generateId, err)
		return err
	}

	logs.Infof("successfully sync device info to cc, subOrderID: %s, generateId: %d, ips: %+v, assets: %+v, "+
		"devices: %+v", order.SubOrderId, generateId, ips, assetIds, cvt.PtrToSlice(devices))

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

func (g *Generator) getHostDetail(assetIds []string) ([]*cmdb.Host, error) {
	req := &cmdb.ListBizHostParams{
		BizID: 931,
		BkModuleIDs: []int64{
			// RA池
			239148,
			// SA云化池
			239149,
			// SCR_加工池
			532040,
		},
		HostPropertyFilter: &cmdb.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    "bk_asset_id",
						Operator: querybuilder.OperatorIn,
						Value:    assetIds,
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
			Limit: pkg.BKMaxInstanceLimit,
		},
	}

	resp, err := g.cc.ListBizHost(kit.New(), req)
	if err != nil {
		logs.Errorf("failed to get cc host info, err: %v", err)
		return nil, err
	}

	hosts := make([]*cmdb.Host, 0)
	for _, host := range resp.Info {
		hosts = append(hosts, cvt.ValToPtr(host))
	}

	return hosts, nil
}

// MatchCVM manual match cvm devices
func (g *Generator) MatchCVM(kt *kit.Kit, param *types.MatchDeviceReq) error {
	// 1. get order by suborder id
	order, err := g.GetApplyOrder(param.SuborderId)
	if err != nil {
		logs.Errorf("failed to match cvm when get apply order, err: %v, order id: %s, rid: %s", err, param.SuborderId,
			kt.Rid)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, param.SuborderId)
	}

	// cannot match device if its stage is not SUSPEND
	if order.Stage != types.TicketStageSuspend {
		logs.Errorf("cannot match device, for order %s stage %s != %s, rid: %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend, kt.Rid)
		return fmt.Errorf("cannot match device, for order %s stage %s != %s", order.SubOrderId, order.Stage,
			types.TicketStageSuspend)
	}

	// 如果是滚服类型，需要进行当月滚服额度的扣减
	if order.RequireType == enumor.RequireTypeRollServer {
		if err = g.rsLogics.ReduceRollingCvmProdAppliedRecord(kt, param.Device); err != nil {
			logs.Errorf("reduce rolling server cvm product applied record failed, err: %+v, devices: %+v, rid: %s", err,
				param.Device, kt.Rid)
			return err
		}
	}

	// set apply order status MATCHING
	if err = g.lockApplyOrder(order); err != nil {
		logs.Errorf("failed to match cvm when lock apply order, err: %v, order id: %s, rid: %s", err, param.SuborderId,
			kt.Rid)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, param.SuborderId)
	}

	replicas := uint(len(param.Device))

	// 2. init generate record
	generateId, err := g.initGenerateRecord(order.ResourceType, order.SubOrderId, replicas, true)
	if err != nil {
		logs.Errorf("failed to match cvm when init generate record, err: %v, order id: %s, rid: %s", err,
			order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, order.SubOrderId)
	}

	// TODO: check whether device is locked by other orders
	deviceList := make([]*types.DeviceInfo, 0)
	successIps := make([]string, 0)
	for _, host := range param.Device {
		deviceList = append(deviceList, &types.DeviceInfo{
			Ip:              host.Ip,
			AssetId:         host.AssetId,
			Deliverer:       param.Operator,
			IsManualMatched: true,
		})
		successIps = append(successIps, host.Ip)
	}

	// 3. save generated cvm instances info
	if err = g.createGeneratedDevice(order, generateId, deviceList); err != nil {
		logs.Errorf("failed to update generated device, err: %v, order id: %s", err, order.SubOrderId)
		return fmt.Errorf("failed to update generated device, err: %v, order id: %s, rid: %s", err, order.SubOrderId,
			kt.Rid)
	}

	// 4. update generate record status to success
	msg := fmt.Sprintf("manually matched by %s successfully", param.Operator)
	if err = g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateId, types.GenerateStatusSuccess,
		msg, "", successIps); err != nil {
		logs.Errorf("failed to match cvm when update generate record, err: %v, order id: %s, rid: %s", err,
			order.SubOrderId, kt.Rid)
		return fmt.Errorf("failed to match cvm, err: %v, order id: %s", err, order.SubOrderId)
	}

	return nil
}

// GetApplyOrder gets apply order by order id
func (g *Generator) GetApplyOrder(key string) (*types.ApplyOrder, error) {
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
	order, err := g.GetApplyOrder(param.SuborderId)
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
