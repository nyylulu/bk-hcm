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

package cvm

import (
	"fmt"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
)

const (
	ApplyLimit = 1000
)

// CvmCreateReq create cvm request
type CvmCreateReq struct {
	BkBizId     int64      `json:"bk_biz_id"`
	BkModuleId  int64      `json:"bk_module_id"`
	User        string     `json:"bk_username"`
	RequireType int64      `json:"require_type"`
	Replicas    uint       `json:"replicas"`
	Remark      string     `json:"remark"`
	Spec        *OrderSpec `json:"spec" bson:"spec"`
}

// Validate whether CvmCreateReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (s *CvmCreateReq) Validate() (errKey string, err error) {

	if s.Replicas <= 0 {
		return "replicas", fmt.Errorf("invalid replicas <= 0")
	}
	// replicas limit 1000
	if s.Replicas > ApplyLimit {
		return "replicas", fmt.Errorf("exceed apply limit: %d", ApplyLimit)
	}

	remarkLimit := 256
	if len(s.Remark) > remarkLimit {
		return "remark", fmt.Errorf("exceed size limit %d", remarkLimit)
	}

	if key, err := s.Spec.Validate(); err != nil {
		return fmt.Sprintf("spec.%s", key), err
	}

	return "", nil
}

// OrderSpec cvm apply order specification
type OrderSpec struct {
	Region      string `json:"region" bson:"region"`
	Zone        string `json:"zone" bson:"zone"`
	DeviceType  string `json:"device_type" bson:"device_type"`
	ImageId     string `json:"image_id" bson:"image_id"`
	DiskSize    int64  `json:"disk_size" bson:"disk_size"`
	DiskType    string `json:"disk_type" bson:"disk_type"`
	NetworkType string `json:"network_type" bson:"network_type"`
	Vpc         string `json:"vpc" bson:"vpc"`
	Subnet      string `json:"subnet" bson:"subnet"`
}

// Validate whether OrderSpec is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (s *OrderSpec) Validate() (errKey string, err error) {
	if len(s.Region) == 0 {
		return "region", fmt.Errorf("cannot be empty")
	}

	if len(s.Vpc) > 0 && len(s.Subnet) == 0 {
		return "subnet", fmt.Errorf("cannot be empty while vpc is set")
	}

	if s.DiskSize < 0 {
		return "disk_size", fmt.Errorf("invalid value < 0")
	}

	diskLimit := int64(16000)
	if s.DiskSize > diskLimit {
		return "disk_size", fmt.Errorf("exceed limit %d", diskLimit)
	}

	// 规格为 10 的倍数
	diskUnit := int64(10)
	modDisk := s.DiskSize % diskUnit
	if modDisk != 0 {
		return "disk_size", fmt.Errorf("must be in multiples of %d", diskUnit)
	}

	return "", nil
}

// CvmCreateResult result of create cvm order
type CvmCreateResult struct {
	OrderId uint64 `json:"order_id"`
}

// ApplyOrder cvm apply order
type ApplyOrder struct {
	OrderId     uint64      `json:"order_id" bson:"order_id"`
	BkBizId     int64       `json:"bk_biz_id" bson:"bk_biz_id"`
	BkModuleId  int64       `json:"bk_module_id" bson:"bk_module_id"`
	User        string      `json:"bk_username" bson:"bk_username"`
	RequireType int64       `json:"require_type" bson:"require_type"`
	Remark      string      `json:"remark" bson:"remark"`
	Spec        *OrderSpec  `json:"spec" bson:"spec"`
	Status      ApplyStatus `json:"status" bson:"status"`
	Message     string      `json:"message" bson:"message"`
	TaskId      string      `json:"task_id" bson:"task_id"`
	TaskLink    string      `json:"task_link" bson:"task_link"`
	Total       uint        `json:"total_num" bson:"total_num"`
	SuccessNum  uint        `json:"success_num" bson:"success_num"`
	FailedNum   uint        `json:"failed_num" bson:"failed_num"`
	PendingNum  uint        `json:"pending_num" bson:"pending_num"`
	CreateAt    time.Time   `json:"create_at" bson:"create_at"`
	UpdateAt    time.Time   `json:"update_at" bson:"update_at"`
}

type ApplyStatus string

const (
	ApplyStatusInit     ApplyStatus = "INIT"
	ApplyStatusRunning  ApplyStatus = "RUNNING"
	RecycleStatusPaused ApplyStatus = "PAUSED"
	RecycleStatusDone   ApplyStatus = "DONE"
	ApplyStatusSuccess  ApplyStatus = "SUCCESS"
	ApplyStatusFailed   ApplyStatus = "FAILED"
)

// GetApplyParam get apply order request parameter
type GetApplyParam struct {
	OrderId     []uint64          `json:"order_id" bson:"order_id"`
	TaskId      []string          `json:"task_id" bson:"task_id"`
	User        []string          `json:"bk_username" bson:"bk_username"`
	RequireType []int64           `json:"require_type" bson:"require_type"`
	Status      []ApplyStatus     `json:"status" bson:"status"`
	Region      []string          `json:"region" bson:"region"`
	Zone        []string          `json:"zone" bson:"zone"`
	DeviceType  []string          `json:"device_type" bson:"device_type"`
	Start       string            `json:"start" bson:"start"`
	End         string            `json:"end" bson:"end"`
	Page        metadata.BasePage `json:"page" bson:"page"`
}

// Validate whether GetApplyParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetApplyParam) Validate() (errKey string, err error) {
	arrayLimit := 20
	if len(param.OrderId) > arrayLimit {
		return "order_id", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.TaskId) > arrayLimit {
		return "task_id", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return "bk_username", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.RequireType) > arrayLimit {
		return "require_type", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.Status) > arrayLimit {
		return "status", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.Region) > arrayLimit {
		return "region", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.Zone) > arrayLimit {
		return "zone", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.DeviceType) > arrayLimit {
		return "device_type", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	return "", nil
}

const (
	dateLayout     = "2006-01-02"
	datetimeLayout = "2006-01-02 15:04:05"
)

// GetFilter get mgo filter
func (param *GetApplyParam) GetFilter() map[string]interface{} {
	filter := make(map[string]interface{})
	if len(param.OrderId) > 0 {
		filter["order_id"] = mapstr.MapStr{
			common.BKDBIN: param.OrderId,
		}
	}
	if len(param.TaskId) > 0 {
		filter["task_id"] = mapstr.MapStr{
			common.BKDBIN: param.TaskId,
		}
	}
	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			common.BKDBIN: param.User,
		}
	}
	if len(param.RequireType) > 0 {
		filter["require_type"] = mapstr.MapStr{
			common.BKDBIN: param.RequireType,
		}
	}
	if len(param.Status) > 0 {
		filter["status"] = mapstr.MapStr{
			common.BKDBIN: param.Status,
		}
	}
	if len(param.Region) > 0 {
		filter["spec.region"] = mapstr.MapStr{
			common.BKDBIN: param.Region,
		}
	}
	if len(param.Zone) > 0 {
		filter["spec.zone"] = mapstr.MapStr{
			common.BKDBIN: param.Zone,
		}
	}
	if len(param.DeviceType) > 0 {
		filter["spec.device_type"] = mapstr.MapStr{
			common.BKDBIN: param.DeviceType,
		}
	}
	timeCond := make(map[string]interface{})
	if len(param.Start) != 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[common.BKDBGTE] = startTime
		}
	}
	if len(param.End) != 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[common.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}
	if len(timeCond) != 0 {
		filter["create_at"] = timeCond
	}

	return filter
}

// CvmInfo cvm device info
type CvmInfo struct {
	OrderId   uint64    `json:"order_id" bson:"order_id"`
	CvmTaskId string    `json:"cvm_task_id" bson:"cvm_task_id"`
	CvmInstId string    `json:"cvm_inst_id" bson:"cvm_inst_id"`
	AssetId   string    `json:"asset_id" bson:"asset_id"`
	Ip        string    `json:"ip" bson:"ip"`
	UpdateAt  time.Time `json:"update_at" bson:"update_at"`
}

// CvmOrderReq cvm apply order query request
type CvmOrderReq struct {
	OrderId int64 `json:"order_id"`
}

// CvmOrderResult cvm apply order query result
type CvmOrderResult struct {
	Count int64         `json:"count"`
	Info  []*ApplyOrder `json:"info"`
}

// CvmDeviceReq cvm apply order launched instances query request
type CvmDeviceReq struct {
	OrderId int64 `json:"order_id"`
}

// CvmDeviceResult cvm apply order launched instances query request
type CvmDeviceResult struct {
	Count int64      `json:"count"`
	Info  []*CvmInfo `json:"info"`
}

// CvmCapacityReq cvm apply capacity query request
type CvmCapacityReq struct {
	BkBizId     int64  `json:"bk_biz_id"`
	RequireType int64  `json:"require_type"`
	Region      string `json:"region"`
	Zone        string `json:"zone"`
	VpcId       string `json:"vpc_id"`
	SubnetId    string `json:"subnet_id"`
	DeviceType  string `json:"device_type"`
}

// CvmCapacityResult cvm apply capacity query result
type CvmCapacityResult struct {
	Count int64           `json:"count"`
	Info  []*CapacityItem `json:"info"`
}

// CapacityItem cvm apply capacity item
type CapacityItem struct {
	Region   string          `json:"region"`
	Zone     string          `json:"zone"`
	VpcId    string          `json:"vpc_id"`
	SubnetId string          `json:"subnet_id"`
	MaxNum   int             `json:"max_num"`
	MaxInfo  []*CapacityInfo `json:"max_info"`
}

// CapacityInfo cvm apply capacity into
type CapacityInfo struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}
