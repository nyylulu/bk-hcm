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

// Package cvm cvm types
package cvm

import (
	"fmt"
	"time"

	"hcm/pkg"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/thirdparty/cvmapi"
	"hcm/pkg/tools/metadata"
)

const (
	// ApplyLimit 申请配额限制
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
func (s *CvmCreateReq) Validate() error {

	if s.Replicas <= 0 {
		return fmt.Errorf("replicas invalid replicas <= 0")
	}
	// replicas limit 1000
	if s.Replicas > ApplyLimit {
		return fmt.Errorf("replicas exceed apply limit: %d", ApplyLimit)
	}

	remarkLimit := 256
	if len(s.Remark) > remarkLimit {
		return fmt.Errorf("remark exceed size limit %d", remarkLimit)
	}

	if err := s.Spec.Validate(); err != nil {
		return err
	}

	return nil
}

// OrderSpec cvm apply order specification
type OrderSpec struct {
	Region      string `json:"region" bson:"region"`
	Zone        string `json:"zone" bson:"zone"`
	DeviceType  string `json:"device_type" bson:"device_type"`
	ImageId     string `json:"image_id" bson:"image_id"`
	NetworkType string `json:"network_type" bson:"network_type"`
	Vpc         string `json:"vpc" bson:"vpc"`
	Subnet      string `json:"subnet" bson:"subnet"`
	// 计费模式(计费模式：PREPAID包年包月，POSTPAID_BY_HOUR按量计费，默认为：PREPAID)
	ChargeType cvmapi.ChargeType `json:"charge_type" bson:"charge_type"`
	// 计费时长，单位：月
	ChargeMonths uint `json:"charge_months" bson:"charge_months"`
	// 被继承云主机实例ID
	InheritInstanceId string            `json:"inherit_instance_id" bson:"inherit_instance_id"`
	SystemDisk        enumor.DiskSpec   `json:"system_disk" bson:"system_disk"`
	DataDisk          []enumor.DiskSpec `json:"data_disk" bson:"data_disk"`
}

// Validate whether OrderSpec is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (s *OrderSpec) Validate() error {
	if len(s.Region) == 0 {
		return fmt.Errorf("region cannot be empty")
	}

	if len(s.Vpc) > 0 && len(s.Subnet) == 0 {
		return fmt.Errorf("subnet cannot be empty while vpc is set")
	}

	// 计费模式校验-该接口没有对外提供，可以为必传参数
	if err := s.ChargeType.Validate(); err != nil {
		return err
	}

	// 包年包月时，计费时长必传
	if s.ChargeType == cvmapi.ChargeTypePrePaid && s.ChargeMonths < 1 {
		return fmt.Errorf("charge_months invalid value < 1")
	}

	// 系统盘类型校验
	if err := s.SystemDisk.Validate(); err != nil {
		return err
	}
	if s.SystemDisk.DiskSize < 50 || s.SystemDisk.DiskSize > 1000 {
		return fmt.Errorf("system_disk_size invalid value, must be in range [50, 1000]")
	}
	// 系统盘大小必须是50的倍数
	if s.SystemDisk.DiskSize%50 != 0 {
		return fmt.Errorf("system_disk_size must be a multiple of 50")
	}

	// 数据盘类型校验
	if len(s.DataDisk) > 0 {
		dataDiskTotalNum := 0
		for _, dd := range s.DataDisk {
			if err := dd.Validate(); err != nil {
				return err
			}
			if dd.DiskSize < 10 || dd.DiskSize > 32000 {
				return fmt.Errorf("data_disk_size invalid value, must be in range [10, 32000]")
			}
			// 数据盘大小必须是10的倍数
			if dd.DiskSize%10 != 0 {
				return fmt.Errorf("data_disk_size must be a multiple of 10")
			}
			dataDiskTotalNum += dd.DiskNum
		}
		// 数据盘总数量不能超过20块
		if dataDiskTotalNum < 0 || dataDiskTotalNum > 20 {
			return fmt.Errorf("data_disk_total_num invalid value, must be in range [0, 20]")
		}
	}

	return nil
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

// ApplyStatus cvm apply order status
type ApplyStatus string

// ApplyStatus cvm apply order status
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
func (param *GetApplyParam) Validate() error {
	arrayLimit := 20
	if len(param.OrderId) > arrayLimit {
		return fmt.Errorf("order_id exceed limit %d", arrayLimit)
	}

	if len(param.TaskId) > arrayLimit {
		return fmt.Errorf("task_id exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return fmt.Errorf("bk_username exceed limit %d", arrayLimit)
	}

	if len(param.RequireType) > arrayLimit {
		return fmt.Errorf("require_type exceed limit %d", arrayLimit)
	}

	if len(param.Status) > arrayLimit {
		return fmt.Errorf("status exceed limit %d", arrayLimit)
	}

	if len(param.Region) > arrayLimit {
		return fmt.Errorf("region exceed limit %d", arrayLimit)
	}

	if len(param.Zone) > arrayLimit {
		return fmt.Errorf("zone exceed limit %d", arrayLimit)
	}

	if len(param.DeviceType) > arrayLimit {
		return fmt.Errorf("device_type exceed limit %d", arrayLimit)
	}

	return nil
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
			pkg.BKDBIN: param.OrderId,
		}
	}
	if len(param.TaskId) > 0 {
		filter["task_id"] = mapstr.MapStr{
			pkg.BKDBIN: param.TaskId,
		}
	}
	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			pkg.BKDBIN: param.User,
		}
	}
	if len(param.RequireType) > 0 {
		filter["require_type"] = mapstr.MapStr{
			pkg.BKDBIN: param.RequireType,
		}
	}
	if len(param.Status) > 0 {
		filter["status"] = mapstr.MapStr{
			pkg.BKDBIN: param.Status,
		}
	}
	if len(param.Region) > 0 {
		filter["spec.region"] = mapstr.MapStr{
			pkg.BKDBIN: param.Region,
		}
	}
	if len(param.Zone) > 0 {
		filter["spec.zone"] = mapstr.MapStr{
			pkg.BKDBIN: param.Zone,
		}
	}
	if len(param.DeviceType) > 0 {
		filter["spec.device_type"] = mapstr.MapStr{
			pkg.BKDBIN: param.DeviceType,
		}
	}
	timeCond := make(map[string]interface{})
	if len(param.Start) != 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}
	if len(param.End) != 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
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
