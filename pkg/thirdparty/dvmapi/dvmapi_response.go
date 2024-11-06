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

package dvmapi

// OrderCreateResp dvm create order response
type OrderCreateResp struct {
	BillId string `json:"billid"`
}

// OrderQueryResp dvm order query response
type OrderQueryResp struct {
	TaskList []TaskList `json:"task_list"`
}

// TaskList dvm order task list
type TaskList struct {
	ID       int64  `json:"id"`
	VMBillId string `json:"billId"`
	IP       string `json:"ip"`
	Status   string `json:"status"`
	Message  string `json:"msg"`
}

// dvm apply order status
const (
	DockerVMSucceeded string = "succ"
	DockerVMFailed    string = "fail"
	DockerVMRunning   string = "running"
	DockerVMWaiting   string = "normal"
)

// DockerCluster docker host Cluster
type DockerCluster struct {
	City                   string `json:"city"`
	SetId                  string `json:"setId"`
	SetName                string `json:"setName"`
	ClusterType            int    `json:"clusterType"`
	IsTlinux2              int    `json:"isTlinux2"`
	IsAutoResourcePlanning int    `json:"isAutoResourcePlanning"`
}

// Host docker host info
type DockerHost struct {
	AppId            string `json:"appId"`
	IP               string `json:"ip"`
	AssetID          string `json:"assetID"`
	DeviceClass      string `json:"deviceClass"`
	Region           string `json:"region"`
	SZone            string `json:"szone"`
	IDCUnitID        string `json:"idcUnitID:"`
	Equipment        string `json:"equipment"`
	ModuleName       string `json:"moduleName"`
	OSVersion        string `json:"osVersion"`
	Memo             string `json:"memo"`
	SetId            string `json:"setId"`
	ScheduledVMs     int    `json:"scheduledVMs"`
	CPUCapacity      int    `json:"cpuCapacity"`
	MemoryCapacity   int    `json:"memoryCapacity"`
	DiskCapacity     int    `json:"diskCapacity"`
	PodCapacity      int    `json:"podCapacity"`
	AllocatableCount int    `json:"allocatableCount"`
	AllocatableCPU   int    `json:"allocatableCpu"`
	AllocatableMem   int    `json:"allocatableMem"`
	AllocatableDisk  int    `json:"allocatableDisk"`
	// Score for prioritize
	Score float64 `json:"score"`
}
