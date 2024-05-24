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

// Package pool defines various data type for service
package pool

import (
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/dal/pool/table"
)

const (
	dateLayout     = "2006-01-02"
	datetimeLayout = "2006-01-02 15:04:05"
)

// ResourceType resource type
type ResourceType string

// ResourceType resource type
const (
	ResourceTypePm          ResourceType = "IDCPM"
	ResourceTypeCvm         ResourceType = "QCLOUDCVM"
	ResourceTypeIdcDvm      ResourceType = "IDCDVM"
	ResourceTypeQcloudDvm   ResourceType = "QCLOUDDVM"
	ResourceTypeOthers      ResourceType = "OTHERS"
	ResourceTypeUnsupported ResourceType = "UNSUPPORTED"
)

// GetLaunchMatchDeviceReq get resource launch match devices request
type GetLaunchMatchDeviceReq struct {
	ResourceType ResourceType `json:"resource_type"`
	Ips          []string     `json:"ips"`
	AssetIDs     []string     `json:"asset_ids"`
	Spec         *MatchSpec   `json:"spec"`
}

// MatchSpec resource launch match specification
type MatchSpec struct {
	Region      []string `json:"region"`
	Zone        []string `json:"zone"`
	DeviceType  []string `json:"device_type"`
	Image       []string `json:"image"`
	OsType      []string `json:"os_type"`
	RaidType    []string `json:"raid_type"`
	DiskType    []string `json:"disk_type"`
	NetworkType []string `json:"network_type"`
	Isp         []string `json:"isp"`
}

// Validate whether GetLaunchMatchDeviceReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetLaunchMatchDeviceReq) Validate() (errKey string, err error) {
	if len(param.Ips) > common.BKMaxInstanceLimit {
		return "ips", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	if len(param.AssetIDs) > common.BKMaxInstanceLimit {
		return "asset_ids", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// GetLaunchMatchDeviceRst get resource launch match devices result
type GetLaunchMatchDeviceRst struct {
	Count int64          `json:"count"`
	Info  []*MatchDevice `json:"info"`
}

// MatchDevice resource launch match device info
type MatchDevice struct {
	BkHostId     int64  `json:"bk_host_id"`
	AssetId      string `json:"asset_id"`
	Ip           string `json:"ip"`
	OuterIp      string `json:"outer_ip"`
	Isp          string `json:"isp"`
	DeviceType   string `json:"device_type"`
	OsType       string `json:"os_type"`
	Region       string `json:"region"`
	Zone         string `json:"zone"`
	Module       string `json:"module"`
	Equipment    int64  `json:"equipment"`
	IdcUnit      string `json:"idc_unit"`
	IdcLogicArea string `json:"idc_logic_area"`
	RaidType     string `json:"raid_type"`
	InputTime    string `json:"input_time"`
}

// GetRecallMatchDeviceReq get resource recall match devices request
type GetRecallMatchDeviceReq struct {
	ResourceType ResourceType `json:"resource_type"`
	Spec         *MatchSpec   `json:"spec"`
}

// Validate whether GetRecallMatchDeviceReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecallMatchDeviceReq) Validate() (errKey string, err error) {
	arrayLimit := 20
	if param.Spec != nil {
		if len(param.Spec.DeviceType) > arrayLimit {
			return "spec.device_type", fmt.Errorf("exceed limit %d", arrayLimit)
		}

		if len(param.Spec.Region) > arrayLimit {
			return "spec.region", fmt.Errorf("exceed limit %d", arrayLimit)
		}

		if len(param.Spec.Zone) > arrayLimit {
			return "spec.zone", fmt.Errorf("exceed limit %d", arrayLimit)
		}
	}

	return "", nil
}

// GetRecallMatchDeviceRst get resource recall match devices result
type GetRecallMatchDeviceRst struct {
	Count int64                `json:"count"`
	Info  []*RecallMatchDevice `json:"info"`
}

// RecallMatchDevice resource recall match device info
type RecallMatchDevice struct {
	DeviceType string `json:"device_type"`
	Region     string `json:"region"`
	Zone       string `json:"zone"`
	Amount     int    `json:"amount"`
}

// LaunchReq create resource launch task request
type LaunchReq struct {
	HostIDs []int64 `json:"bk_host_ids"`
}

// Validate whether LaunchReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *LaunchReq) Validate() (errKey string, err error) {
	if len(param.HostIDs) == 0 {
		return "bk_host_ids", errors.New("cannot be empty")
	}

	if len(param.HostIDs) > common.BKMaxInstanceLimit {
		return "bk_host_ids", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// RecallReq create resource recall task request
type RecallReq struct {
	DeviceType string  `json:"device_type"`
	Region     string  `json:"region"`
	Zone       string  `json:"zone"`
	AssetIDs   []int64 `json:"asset_ids"`
	Replicas   uint    `json:"replicas"`
}

// Validate whether RecallReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *RecallReq) Validate() (errKey string, err error) {
	if param.DeviceType == "" {
		return "device_type", errors.New("cannot be empty")
	}

	if len(param.AssetIDs) > common.BKMaxInstanceLimit {
		return "asset_ids", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	if param.Replicas <= 0 {
		return "replicas", errors.New("should be positive")
	}

	if param.Replicas > common.BKMaxInstanceLimit {
		return "replicas", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// GetLaunchTaskReq get resource launch task request
type GetLaunchTaskReq struct {
	ID    []uint64            `json:"id"`
	User  []string            `json:"bk_username"`
	Phase []table.OpTaskPhase `json:"phase"`
	Start string              `json:"start"`
	End   string              `json:"end"`
	Page  metadata.BasePage   `json:"page"`
}

// Validate whether GetLaunchTaskReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetLaunchTaskReq) Validate() (errKey string, err error) {
	arrayLimit := 20
	if len(param.ID) > arrayLimit {
		return "id", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.Phase) > arrayLimit {
		return "phase", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return "bk_username", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.Start) > 0 {
		_, err := time.Parse(dateLayout, param.Start)
		if err != nil {
			return "start", fmt.Errorf("date format should be like %s", dateLayout)
		}
	}

	if len(param.End) > 0 {
		_, err := time.Parse(dateLayout, param.End)
		if err != nil {
			return "end", fmt.Errorf("date format should be like %s", dateLayout)
		}
	}

	if key, err := param.Page.Validate(false); err != nil {
		return key, err
	}

	if param.Page.Start < 0 {
		return "page.start", errors.New("invalid start < 0")
	}

	if param.Page.Limit < 0 {
		return "page.limit", errors.New("invalid limit < 0")
	}

	if param.Page.Limit > 200 {
		return "page.limit", errors.New("exceed limit 200")
	}

	return "", nil
}

// GetFilter get mgo filter
func (param *GetLaunchTaskReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.ID) > 0 {
		filter["id"] = mapstr.MapStr{
			common.BKDBIN: param.ID,
		}
	}

	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			common.BKDBIN: param.User,
		}
	}

	if len(param.Phase) > 0 {
		filter["status.phase"] = mapstr.MapStr{
			common.BKDBIN: param.Phase,
		}
	}

	timeCond := make(map[string]interface{})
	if len(param.Start) > 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[common.BKDBGTE] = startTime
		}
	}

	if len(param.End) > 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[common.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}

	if len(timeCond) > 0 {
		filter["create_at"] = timeCond
	}

	return filter, nil
}

// GetLaunchTaskRst get pool launch task result
type GetLaunchTaskRst struct {
	Count int64               `json:"count"`
	Info  []*table.LaunchTask `json:"info"`
}

// GetRecallTaskReq get resource recall task request
type GetRecallTaskReq struct {
	ID    []uint64            `json:"id"`
	User  []string            `json:"bk_username"`
	Phase []table.OpTaskPhase `json:"phase"`
	Start string              `json:"start"`
	End   string              `json:"end"`
	Page  metadata.BasePage   `json:"page"`
}

// Validate whether GetRecallTaskReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecallTaskReq) Validate() (errKey string, err error) {
	arrayLimit := 20
	if len(param.ID) > arrayLimit {
		return "id", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.Phase) > arrayLimit {
		return "phase", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return "bk_username", fmt.Errorf("exceed limit %d", arrayLimit)
	}

	if len(param.Start) > 0 {
		_, err := time.Parse(dateLayout, param.Start)
		if err != nil {
			return "start", fmt.Errorf("date format should be like %s", dateLayout)
		}
	}

	if len(param.End) > 0 {
		_, err := time.Parse(dateLayout, param.End)
		if err != nil {
			return "end", fmt.Errorf("date format should be like %s", dateLayout)
		}
	}

	if key, err := param.Page.Validate(false); err != nil {
		return key, err
	}

	if param.Page.Start < 0 {
		return "page.start", errors.New("invalid start < 0")
	}

	if param.Page.Limit < 0 {
		return "page.limit", errors.New("invalid limit < 0")
	}

	if param.Page.Limit > 200 {
		return "page.limit", errors.New("exceed limit 200")
	}

	return "", nil
}

// GetFilter get mgo filter
func (param *GetRecallTaskReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.ID) > 0 {
		filter["id"] = mapstr.MapStr{
			common.BKDBIN: param.ID,
		}
	}

	if len(param.Phase) > 0 {
		filter["status.phase"] = mapstr.MapStr{
			common.BKDBIN: param.Phase,
		}
	}

	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			common.BKDBIN: param.User,
		}
	}

	timeCond := make(map[string]interface{})
	if len(param.Start) > 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[common.BKDBGTE] = startTime
		}
	}

	if len(param.End) > 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[common.BKDBLT] = endTime.AddDate(0, 0, 1)
		}
	}

	if len(timeCond) > 0 {
		filter["create_at"] = timeCond
	}

	return filter, nil
}

// GetRecallTaskRst get pool launch recall result
type GetRecallTaskRst struct {
	Count int64               `json:"count"`
	Info  []*table.RecallTask `json:"info"`
}

// GetLaunchHostReq get resource launch host request
type GetLaunchHostReq struct {
	ID     uint64                    `json:"id"`
	Filter *querybuilder.QueryFilter `json:"filter"`
	Page   metadata.BasePage         `json:"page"`
}

// Validate whether GetLaunchHostReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetLaunchHostReq) Validate() (errKey string, err error) {
	if param.ID <= 0 {
		return "id", errors.New("should be positive")
	}

	if key, err := param.Page.Validate(false); err != nil {
		return key, err
	}

	if param.Page.Start < 0 {
		return "page.start", errors.New("invalid start < 0")
	}

	if param.Page.Limit < 0 {
		return "page.limit", errors.New("invalid limit < 0")
	}

	if param.Page.Limit > common.BKMaxInstanceLimit {
		return "page.limit", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	if param.Filter != nil {
		if key, err := param.Filter.Validate(&querybuilder.RuleOption{NeedSameSliceElementType: true}); err != nil {
			return fmt.Sprintf("filter.%s", key), err
		}
		if param.Filter.GetDeep() > querybuilder.MaxDeep {
			return "filter.rules", fmt.Errorf("exceed max query condition deepth: %d",
				querybuilder.MaxDeep)
		}
	}

	return "", nil
}

// GetFilter get mgo filter
func (param *GetLaunchHostReq) GetFilter() (map[string]interface{}, error) {
	if param.Filter != nil {
		mgoFilter, key, err := param.Filter.ToMgo()
		if err != nil {
			return nil, fmt.Errorf("invalid key:filter.%s, err: %s", key, err)
		}
		mgoFilter["op_type"] = table.OpTypeLaunch
		mgoFilter["task_id"] = param.ID
		return mgoFilter, nil
	}

	filter := map[string]interface{}{
		"op_type": table.OpTypeLaunch,
		"task_id": param.ID,
	}

	return filter, nil
}

// GetLaunchHostRst get pool launch host result
type GetLaunchHostRst struct {
	Count int64             `json:"count"`
	Info  []*table.OpRecord `json:"info"`
}

// GetRecallHostReq get resource recall host request
type GetRecallHostReq struct {
	ID   uint64            `json:"id"`
	Page metadata.BasePage `json:"page"`
}

// Validate whether GetRecallHostReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecallHostReq) Validate() (errKey string, err error) {
	if param.ID < 0 {
		return "id", errors.New("cannot be negative")
	}

	if key, err := param.Page.Validate(false); err != nil {
		return key, err
	}

	if param.Page.Start < 0 {
		return "page.start", errors.New("invalid start < 0")
	}

	if param.Page.Limit < 0 {
		return "page.limit", errors.New("invalid limit < 0")
	}

	if param.Page.Limit > common.BKMaxInstanceLimit {
		return "page.limit", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// GetRecallHostRst get pool launch host result
type GetRecallHostRst struct {
	Count int64             `json:"count"`
	Info  []*table.OpRecord `json:"info"`
}

// GetPoolHostReq get resource pool host request
type GetPoolHostReq struct {
	Selector []*table.Selector     `json:"selector"`
	Phase    []table.PoolHostPhase `json:"phase"`
	Page     metadata.BasePage     `json:"page"`
}

// Validate whether GetPoolHostReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetPoolHostReq) Validate() (errKey string, err error) {
	for idx, selector := range param.Selector {
		if key, err := selector.Validate(); err != nil {
			return fmt.Sprintf("selector[%d].%s", idx, key), err
		}
	}

	if key, err := param.Page.Validate(false); err != nil {
		return key, err
	}

	if param.Page.Start < 0 {
		return "page.start", errors.New("invalid start < 0")
	}

	if param.Page.Limit < 0 {
		return "page.limit", errors.New("invalid limit < 0")
	}

	if param.Page.Limit > common.BKMaxInstanceLimit {
		return "page.limit", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// GetFilter get mgo filter
func (param *GetPoolHostReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.Phase) > 0 {
		filter["status.phase"] = mapstr.MapStr{
			common.BKDBIN: param.Phase,
		}
	}

	for _, selector := range param.Selector {
		key := fmt.Sprintf("labels.%s", selector.Key)
		switch selector.Operator {
		case table.SelectOpEqual:
			filter[key] = selector.Value
		case table.SelectOpIn:
			filter[key] = mapstr.MapStr{
				common.BKDBIN: selector.Value,
			}
		}
	}

	return filter, nil
}

// GetPoolHostRst get pool host result
type GetPoolHostRst struct {
	Count int64             `json:"count"`
	Info  []*table.PoolHost `json:"info"`
}

// DrawHostReq draw hosts from resource pool request
type DrawHostReq struct {
	HostIDs []int64 `json:"bk_host_ids"`
	ToBizID int64   `json:"bk_biz_id"`
}

// Validate whether DrawHostReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *DrawHostReq) Validate() (errKey string, err error) {
	if len(param.HostIDs) == 0 {
		return "bk_host_ids", errors.New("cannot be empty")
	}

	if len(param.HostIDs) > common.BKMaxInstanceLimit {
		return "bk_host_ids", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	if param.ToBizID <= 0 {
		return "bk_biz_id", errors.New("should be positive")
	}

	return "", nil
}

// ReturnHostReq return hosts from resource pool request
type ReturnHostReq struct {
	RecallID  uint64  `json:"recall_id"`
	FromBizID int64   `json:"bk_biz_id"`
	HostIDs   []int64 `json:"bk_host_ids"`
}

// Validate whether ReturnHostReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *ReturnHostReq) Validate() (errKey string, err error) {
	if param.RecallID <= 0 {
		return "recall_id", errors.New("should be positive")
	}

	if param.FromBizID <= 0 {
		return "bk_biz_id", errors.New("should be positive")
	}

	if len(param.HostIDs) == 0 {
		return "bk_host_ids", errors.New("cannot be empty")
	}

	if len(param.HostIDs) > common.BKMaxInstanceLimit {
		return "bk_host_ids", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// CreateRecallOrderReq create resource recall order request
type CreateRecallOrderReq struct {
	DeviceType string  `json:"device_type"`
	Region     string  `json:"region"`
	Zone       string  `json:"zone"`
	AssetIDs   []int64 `json:"asset_ids"`
	ImageID    string  `json:"image_id"`
	OsType     string  `json:"os_type"`
	Replicas   uint    `json:"replicas"`
}

// Validate whether CreateRecallOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *CreateRecallOrderReq) Validate() (errKey string, err error) {
	if param.DeviceType == "" {
		return "device_type", errors.New("cannot be empty")
	}

	if len(param.AssetIDs) > common.BKMaxInstanceLimit {
		return "asset_ids", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	if param.Replicas <= 0 {
		return "replicas", errors.New("should be positive")
	}

	if param.Replicas > common.BKMaxInstanceLimit {
		return "replicas", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// GetRecallOrderReq get resource recall order request
type GetRecallOrderReq struct {
	ID uint64 `json:"id"`
}

// Validate whether GetRecallOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecallOrderReq) Validate() (errKey string, err error) {
	if param.ID <= 0 {
		return "id", errors.New("should be positive")
	}

	return "", nil
}

// GetRecallOrderRst get pool recall order result
type GetRecallOrderRst struct {
	Count int64                `json:"count"`
	Info  []*table.RecallOrder `json:"info"`
}

// GetRecalledInstReq get resource recalled instance request
type GetRecalledInstReq struct {
	ID uint64 `json:"id"`
}

// Validate whether GetRecalledInstReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecalledInstReq) Validate() (errKey string, err error) {
	if param.ID <= 0 {
		return "id", errors.New("should be positive")
	}

	return "", nil
}

// GetRecalledInstRst get pool recalled instance result
type GetRecalledInstRst struct {
	Count int64                 `json:"count"`
	Info  []*table.RecallDetail `json:"info"`
}

// GetRecallDetailReq get recall task detail info request
type GetRecallDetailReq struct {
	ID   uint64            `json:"id"`
	Page metadata.BasePage `json:"page"`
}

// Validate whether GetRecallDetailReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecallDetailReq) Validate() (errKey string, err error) {
	if param.ID <= 0 {
		return "id", errors.New("should be positive")
	}

	if key, err := param.Page.Validate(false); err != nil {
		return key, err
	}

	if param.Page.Start < 0 {
		return "page.start", errors.New("invalid start < 0")
	}

	if param.Page.Limit < 0 {
		return "page.limit", errors.New("invalid limit < 0")
	}

	if param.Page.Limit > common.BKMaxInstanceLimit {
		return "page.limit", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// GetRecallDetailRst get pool recall task detail info result
type GetRecallDetailRst struct {
	Count int64                 `json:"count"`
	Info  []*table.RecallDetail `json:"info"`
}

// ResumeRecycleTaskReq resume recycle task request
type ResumeRecycleTaskReq struct {
	ID []string `json:"id"`
}

// Validate whether ResumeRecycleTaskReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *ResumeRecycleTaskReq) Validate() (errKey string, err error) {
	if len(param.ID) == 0 {
		return "id", fmt.Errorf("id should be set")
	}

	if len(param.ID) > common.BKMaxInstanceLimit {
		return "id", fmt.Errorf("exceed limit %d", common.BKMaxInstanceLimit)
	}

	return "", nil
}

// GetGradeCfgRst get pool grade config result
type GetGradeCfgRst struct {
	Info []*table.GradeCfg `json:"info"`
}

// GetDeviceTypeRst get pool supported device type result
type GetDeviceTypeRst struct {
	Info []interface{} `json:"info"`
}
