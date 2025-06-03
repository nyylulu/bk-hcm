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
	"fmt"
	"sort"
	"strings"
	"sync"

	"hcm/cmd/woa-server/logics/task/scheduler/algorithm"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/dvmapi"
	"hcm/pkg/tools/utils"
)

// createDVM starts a docker vm creating task
func (g *Generator) createDVM(applyRequest *types.DVMSelector, order *types.ApplyOrder,
	host *types.HostPriority, replicas uint) (string, error) {

	req := &dvmapi.OrderCreateReq{
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

	// call dvm api to launchCvm dvm order
	resp, err := g.dvm.CreateDvmOrder(nil, nil, req)
	if err != nil {
		return "", err
	}

	if resp.BillId == "" {
		return "", fmt.Errorf("docker vm order create task return empty order id")
	}

	return resp.BillId, nil
}

// checkDVM checks docker vm creating task result
func (g *Generator) checkDVM(orderId string) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to query dvm order by id %s, err: %v", orderId, err)
		}

		if obj == nil {
			return false, fmt.Errorf("dvm order %s not found", orderId)
		}

		resp, ok := obj.(*dvmapi.OrderQueryResp)
		if !ok {
			return false, fmt.Errorf("object with order id %s is not a dvm order response: %+v", orderId, obj)
		}

		taskList := resp.TaskList
		if len(taskList) == 0 {
			return false, fmt.Errorf("checking docker task (billId = %s)", orderId)
		}

		doingCnt := 0
		for _, item := range taskList {
			if item.Status == dvmapi.DockerVMRunning ||
				item.Status == dvmapi.DockerVMWaiting {
				doingCnt++
				continue
			}
			// 异常状态
			// 有可能生产过程中IP还没有获取到
			// 在下一次检测中处理
			if item.IP == "" && item.Status == dvmapi.DockerVMSucceeded {
				logs.Errorf("task %s IP is empty: %+v", orderId, item)
				doingCnt++
				continue
			}
		}
		if doingCnt > 0 {
			return false, fmt.Errorf("checking docker task (billId = %s)", orderId)
		}
		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// call dvm api to query dvm order status
		return g.dvm.QueryDvmOrders(nil, nil, orderId)
	}

	// TODO: get retry strategy from config
	_, err := utils.Retry(doFunc, checkFunc, 3600, 5)
	return err
}

// listCVM lists created docker vm by order id
func (g *Generator) listDVM(orderId string) ([]dvmapi.TaskList, error) {
	resp, err := g.dvm.QueryDvmOrders(nil, nil, orderId)
	if err != nil {
		return nil, err
	}

	succTasks := make([]dvmapi.TaskList, 0)
	for _, task := range resp.TaskList {
		if task.Status == dvmapi.DockerVMSucceeded && len(task.IP) > 0 {
			succTasks = append(succTasks, task)
		}
	}

	if len(succTasks) <= 0 {
		return nil, fmt.Errorf("no dvm successfully generated")
	}

	return resp.TaskList, nil
}

var hostDeviceClass = map[string]map[string][]string{
	"GAMESERVER": {
		"ONETHOUSAND": []string{"M1", "M10", "M10A", "M10C", "CG1-10G", "M1A", "S3.8XLARGE128", "SN3ne.8XLARGE128",
			"SN3ne.6XLARGE96", "S3.6XLARGE96", "S5.6XLARGE96", "S3NE.6XLARGE96", "S3.6XLARGE160", "SN3ne.6XLARGE160",
			"CG3-10G", "CG2-10G"},
		"TENTHOUSAND": []string{"M10", "M10A", "M10C", "CG1-10G", "M1A", "S3.8XLARGE128", "SN3ne.8XLARGE128",
			"SN3ne.6XLARGE96", "S3.6XLARGE96", "S3NE.6XLARGE96", "S3.6XLARGE160", "SN3ne.6XLARGE160", "CG3-10G",
			"CG2-10G", "S5.8XLARGE128", "S5.6XLARGE96", "S5.4XLARGE128", "SA2.8XLARGE64", "S5.16XLARGE192"},
	},
	"DBSERVICE": {
		"ONETHOUSAND": []string{"Z3", "Z30", "Z30A", "SH3-10G", "SH7-10G", "IT2.6XLARGE192", "IT3a.6XLARGE128",
			"IT3b.6XLARGE128"},
		"TENTHOUSAND": []string{"Z30", "Z30A", "SH3-10G", "SH7-10G", "IT2.6XLARGE192", "IT3a.6XLARGE128",
			"IT3b.6XLARGE128", "IT5.8XLARGE128"},
	},
	"HIGHFREQ": {
		"ONETHOUSAND": []string{},
		"TENTHOUSAND": []string{"Z13"},
	},
}

func (g *Generator) getHostDeviceClassForScheduler(dockerType string, networkType string) []string {
	return hostDeviceClass[dockerType][networkType]
}

func (g *Generator) getSpecialAppRole(appId string) string {
	switch appId {
	// 官网与营销（159）；互娱业务安全（173）；质量开放平台（632）；潘多拉（709）；dc大数据服务集群（100179）；官网AMS道具仓库（100575）
	case "159", "173", "632", "709", "100179", "100575":
		return "PublicPlatform"
	default:
		return ""
	}
}

func (g *Generator) getAMDDevicePattern() []string {
	patterns := []string{"^SA.+"}
	return patterns
}

// filter filters docker hosts according to selector
func (g *Generator) filter(selector *types.DVMSelector, filterHosts []*dvmapi.DockerHost) ([]*dvmapi.DockerHost,
	error) {

	var (
		observedHost    []*dvmapi.DockerHost
		conflictMessage = make(map[string]int)
	)
	for _, host := range filterHosts {
		fit := false
		var err error = nil
		for name, predicate := range g.predicateFuncs {
			fit, err = predicate(selector, host)
			if err != nil || !fit {
				conflictMessage[name]++
				break
			}
		}
		if fit {
			observedHost = append(observedHost, host)
		}
	}
	if len(observedHost) == 0 {
		var parts []string
		for name, val := range conflictMessage {
			parts = append(parts, fmt.Sprintf("%s(%d)", name, val))
		}
		return nil, fmt.Errorf("no hosts are avaliable that match all of the predicates: hosts(%d), %s",
			len(filterHosts), strings.Join(parts, ","))
	}
	return observedHost, nil
}

func (g *Generator) prioritize(selector *types.DVMSelector, hosts []*dvmapi.DockerHost) (types.HostPriorityList,
	error) {

	var (
		mu   = sync.Mutex{}
		wg   = sync.WaitGroup{}
		errs []error
	)
	appendError := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		errs = append(errs, err)
	}
	results := make([]types.HostPriorityList, 0, len(g.priorityFuncs))
	for range g.priorityFuncs {
		results = append(results, nil)
	}
	for i, priorityConfig := range g.priorityFuncs {
		if priorityConfig.Function != nil {
			wg.Add(1)
			go func(index int, config algorithm.PriorityConfig) {
				defer wg.Done()
				var err error
				results[index], err = config.Function(selector, hosts)
				if err != nil {
					appendError(err)
				}
			}(i, priorityConfig)
		} else {
			results[i] = make(types.HostPriorityList, len(hosts))
		}
	}
	wg.Wait()
	if len(errs) != 0 {
		return types.HostPriorityList{}, errs[0]
	}

	result := make(types.HostPriorityList, 0, len(hosts))
	for i, host := range hosts {
		result = append(result, types.HostPriority{
			IP:               host.IP,
			DeviceClass:      host.DeviceClass,
			Equipment:        host.Equipment,
			ModuleName:       host.ModuleName,
			SZone:            host.SZone,
			AllocatableCount: host.AllocatableCount,
			ScheduledVMs:     host.ScheduledVMs,
			SetId:            host.SetId,
			Score:            host.Score,
		})
		for j := range g.priorityFuncs {
			result[i].Score += results[j][i].Score * float64(g.priorityFuncs[j].Weight)
		}
	}
	return result, nil
}

func (g *Generator) sortHosts(antiAffinityLevel string, priorityList types.HostPriorityList) types.HostPriorityList {
	sort.Sort(sort.Reverse(priorityList))
	if antiAffinityLevel == types.AntiNone {
		return priorityList
	}

	hostPriorityList := types.HostPriorityList{}
	tmpHostList := types.HostPriorityList{}
	prefixAntiAffinityValue := ""
	for _, host := range priorityList {
		currentAntiAffinityValue := g.antiAffinityValue(antiAffinityLevel, host)
		if prefixAntiAffinityValue != currentAntiAffinityValue {
			hostPriorityList = append(hostPriorityList, host)
			if len(tmpHostList) > 0 {
				hostPriorityList = append(hostPriorityList, tmpHostList[0])
				tmpHostList = tmpHostList[1:]
			}
		} else {
			tmpHostList = append(tmpHostList, host)
		}
		prefixAntiAffinityValue = currentAntiAffinityValue
	}

	for _, host := range tmpHostList {
		hostPriorityList = append(hostPriorityList, host)
	}

	return hostPriorityList
}

func (g *Generator) antiAffinityValue(antiAffinityLevel string, host types.HostPriority) string {
	var value string
	switch antiAffinityLevel {
	case types.AntiRack:
		value = host.Equipment
	case types.AntiModule:
		value = host.ModuleName
	case types.AntiCampus:
		value = host.SZone
	default:
		value = "none"
	}
	return value
}

func (g *Generator) getAllocatableHosts(kt *kit.Kit, applyRequest *types.DVMSelector, resourceType types.ResourceType,
	existingHostMap map[string]*dvmapi.DockerHost) (types.HostPriorityList, error) {

	// 1. list cluster
	isTlinux2 := g.isTlinux2(applyRequest.Image)
	clusterList, err := g.listDVMCluster(applyRequest.Region, resourceType, isTlinux2)
	if err != nil {
		return nil, err
	}
	logs.V(4).Infof("Cluster list: %v", clusterList)

	// 2. list all docker hosts
	allocatableHosts := make([]*dvmapi.DockerHost, 0)
	for _, cluster := range clusterList {
		hosts, _ := g.listAllocatableHost(cluster, applyRequest)
		if len(hosts) > 0 {
			allocatableHosts = append(allocatableHosts, hosts...)
		}
	}
	if len(allocatableHosts) == 0 {
		logs.Errorf("failed to find any host in cluster")
		return nil, fmt.Errorf("failed to find any host in cluster")
	}

	// 3. 补全Host信息,包括母机对应的虚拟比等
	tmpHosts, err := g.toHost(kt, allocatableHosts, existingHostMap)
	if err != nil {
		logs.Errorf("failed to complete docker host info, err: %v", err)
		return nil, err
	}

	// 4. 预选阶段，筛选满足需求的母机
	observedHosts, err := g.filter(applyRequest, tmpHosts)
	if err != nil {
		logs.Errorf("failed to filter matched docker hosts, err: %v", err)
		return nil, err
	}
	// 优选阶段
	priorityList, err := g.prioritize(applyRequest, observedHosts)
	if err != nil {
		logs.Errorf("failed to sort candidate docker hosts, err: %v", err)
		return nil, err
	}

	return priorityList, nil
}

var regionMappingCHNToCode = map[string]string{
	"天津":  "TJ",
	"深圳":  "SZ",
	"上海":  "SH",
	"香港":  "HK",
	"加拿大": "CA",
	"南京":  "NJ",
	"重庆":  "CQ",
	"广州":  "GZ",
}

const (
	tlinux2ImageName string = "hub.oa.com/library/tlinux2.2"
	maxVirtualRatio  int    = 3
)

func (g *Generator) isTlinux2(imageName string) bool {
	if imageName == "" {
		return false
	}
	parts := strings.Split(imageName, ":")
	return parts[0] == tlinux2ImageName
}

func (g *Generator) listDVMCluster(region string, resourceType types.ResourceType, isTlinux2 bool) (
	[]*dvmapi.DockerCluster, error) {

	rst, err := g.dvm.ListCluster(nil, nil)
	if err != nil {
		return nil, err
	}

	var clusters []*dvmapi.DockerCluster
	for _, cluster := range rst {
		// 如果是腾讯云的Docker只能选腾讯云集群
		if resourceType == types.ResourceTypeQcloudDvm && (cluster.ClusterType != 4 && cluster.ClusterType != 5) {
			continue
		} else if resourceType != types.ResourceTypeQcloudDvm && (cluster.ClusterType == 4 ||
			cluster.ClusterType == 5) {
			continue
		}
		// 匹配tlinux2.2
		if isTlinux2 && cluster.IsTlinux2 != 1 {
			continue
		}
		if cluster.IsAutoResourcePlanning == 1 && cluster.City == regionMappingCHNToCode[region] {
			clusters = append(clusters, cluster)
		}
	}

	return clusters, nil
}

func (g *Generator) listAllocatableHost(cluster *dvmapi.DockerCluster, applyRequest *types.DVMSelector) (
	[]*dvmapi.DockerHost, error) {

	var (
		mu               = sync.Mutex{}
		wg               = sync.WaitGroup{}
		errs             []error
		allocatableHosts []*dvmapi.DockerHost
	)

	appendAllocatableHosts := func(hosts []*dvmapi.DockerHost) {
		mu.Lock()
		defer mu.Unlock()
		for _, host := range hosts {
			host.SetId = cluster.SetId
			allocatableHosts = append(allocatableHosts, host)
		}
	}

	appendError := func(err error) {
		mu.Lock()
		defer mu.Unlock()
		errs = append(errs, err)
	}

	deviceClasses := g.getHostDeviceClassForScheduler(applyRequest.DockerType, applyRequest.NetworkType)
	logs.V(5).Infof("%s %s host device class: %v", applyRequest.DockerType, applyRequest.NetworkType, deviceClasses)
	for _, deviceClass := range deviceClasses {
		wg.Add(1)
		go func(deviceClass string) {
			defer wg.Done()
			req := &dvmapi.ListHostReq{
				SetId:       cluster.SetId,
				DeviceClass: deviceClass,
				Cores:       applyRequest.Cores,
				Memory:      applyRequest.Memory,
				Disk:        applyRequest.Disk,
				HostRole:    applyRequest.HostRole,
			}
			hosts, err := g.dvm.ListHostInCluster(nil, nil, req)
			if err != nil {
				appendError(fmt.Errorf("list host in cluster (%s - %s) failed: %v", cluster.SetId, deviceClass, err))
				return
			}
			appendAllocatableHosts(hosts)
		}(deviceClass)
	}
	wg.Wait()

	if len(errs) > 0 {
		logs.Errorf("failed to get allocatable docker hosts, err: %v", errs[0])
		return nil, errs[0]
	}

	return allocatableHosts, nil
}

func (g *Generator) toHost(kt *kit.Kit, allocatableHosts []*dvmapi.DockerHost,
	existingHostMapping map[string]*dvmapi.DockerHost) ([]*dvmapi.DockerHost, error) {

	hosts := make([]*dvmapi.DockerHost, 0)
	ips := make([]string, 0)
	hostTopoMap := make(map[string]*cmdb.DeviceTopoInfo, 0)

	for _, host := range allocatableHosts {
		ips = append(ips, host.IP)
	}
	if len(ips) > 0 {
		topoInfos, err := g.listDeviceTopo(kt, ips)
		if err != nil {
			return nil, fmt.Errorf("list device topoinfo failed: %v", err)
		}
		for _, topo := range topoInfos {
			hostTopoMap[topo.InnerIP] = topo
		}
	}

	for _, host := range allocatableHosts {
		deviceTopoInfo, exist := hostTopoMap[host.IP]
		if !exist {
			continue
		}
		host.ScheduledVMs = 0
		if existingHost, ok := existingHostMapping[deviceTopoInfo.AssetID]; ok {
			host.ScheduledVMs = existingHost.ScheduledVMs
		}
		host.AssetID = deviceTopoInfo.AssetID
		host.DeviceClass = deviceTopoInfo.DeviceClass
		host.SZone = deviceTopoInfo.SZone
		host.OSVersion = deviceTopoInfo.OSVersion
		host.Equipment = deviceTopoInfo.Equipment
		host.ModuleName = deviceTopoInfo.ModuleName
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func (g *Generator) sumReplicas(replicasMapping map[string]uint) uint {
	sum := uint(0)
	for _, val := range replicasMapping {
		sum += val
	}
	return sum
}
