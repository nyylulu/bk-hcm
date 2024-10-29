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

	"hcm/cmd/woa-server/dal/pool/table"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
	"hcm/pkg/tools/querybuilder"
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
func (param *GetLaunchMatchDeviceReq) Validate() error {
	if len(param.Ips) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("ips exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if len(param.AssetIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("asset_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
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
func (param *GetRecallMatchDeviceReq) Validate() error {
	arrayLimit := 20
	if param.Spec != nil {
		if len(param.Spec.DeviceType) > arrayLimit {
			return fmt.Errorf("spec.device_type exceed limit %d", arrayLimit)
		}

		if len(param.Spec.Region) > arrayLimit {
			return fmt.Errorf("spec.region exceed limit %d", arrayLimit)
		}

		if len(param.Spec.Zone) > arrayLimit {
			return fmt.Errorf("spec.zone exceed limit %d", arrayLimit)
		}
	}

	return nil
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
func (param *LaunchReq) Validate() error {
	if len(param.HostIDs) == 0 {
		return errors.New("bk_host_ids cannot be empty")
	}

	if len(param.HostIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("bk_host_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
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
func (param *RecallReq) Validate() error {
	if param.DeviceType == "" {
		return errors.New("device_type cannot be empty")
	}

	if len(param.AssetIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("asset_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if param.Replicas <= 0 {
		return errors.New("replicas should be positive")
	}

	if param.Replicas > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("replicas exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
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
func (param *GetLaunchTaskReq) Validate() error {
	arrayLimit := 20
	if len(param.ID) > arrayLimit {
		return fmt.Errorf("id exceed limit %d", arrayLimit)
	}

	if len(param.Phase) > arrayLimit {
		return fmt.Errorf("phase exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return fmt.Errorf("bk_username exceed limit %d", arrayLimit)
	}

	if len(param.Start) > 0 {
		_, err := time.Parse(dateLayout, param.Start)
		if err != nil {
			return fmt.Errorf("start date format should be like %s", dateLayout)
		}
	}

	if len(param.End) > 0 {
		_, err := time.Parse(dateLayout, param.End)
		if err != nil {
			return fmt.Errorf("end date format should be like %s", dateLayout)
		}
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return errors.New("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return errors.New("invalid page.limit < 0")
	}

	if param.Page.Limit > 200 {
		return errors.New("exceed page.limit 200")
	}

	return nil
}

// GetFilter get mgo filter
func (param *GetLaunchTaskReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.ID) > 0 {
		filter["id"] = mapstr.MapStr{
			pkg.BKDBIN: param.ID,
		}
	}

	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			pkg.BKDBIN: param.User,
		}
	}

	if len(param.Phase) > 0 {
		filter["status.phase"] = mapstr.MapStr{
			pkg.BKDBIN: param.Phase,
		}
	}

	timeCond := make(map[string]interface{})
	if len(param.Start) > 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}

	if len(param.End) > 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
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
func (param *GetRecallTaskReq) Validate() error {
	arrayLimit := 20
	if len(param.ID) > arrayLimit {
		return fmt.Errorf("id exceed limit %d", arrayLimit)
	}

	if len(param.Phase) > arrayLimit {
		return fmt.Errorf("phase exceed limit %d", arrayLimit)
	}

	if len(param.User) > arrayLimit {
		return fmt.Errorf("bk_username exceed limit %d", arrayLimit)
	}

	if len(param.Start) > 0 {
		_, err := time.Parse(dateLayout, param.Start)
		if err != nil {
			return fmt.Errorf("start date format should be like %s", dateLayout)
		}
	}

	if len(param.End) > 0 {
		_, err := time.Parse(dateLayout, param.End)
		if err != nil {
			return fmt.Errorf("end date format should be like %s", dateLayout)
		}
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return errors.New("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return errors.New("invalid page.limit < 0")
	}

	if param.Page.Limit > 200 {
		return errors.New("exceed page.limit 200")
	}

	return nil
}

// GetFilter get mgo filter
func (param *GetRecallTaskReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.ID) > 0 {
		filter["id"] = mapstr.MapStr{
			pkg.BKDBIN: param.ID,
		}
	}

	if len(param.Phase) > 0 {
		filter["status.phase"] = mapstr.MapStr{
			pkg.BKDBIN: param.Phase,
		}
	}

	if len(param.User) > 0 {
		filter["bk_username"] = mapstr.MapStr{
			pkg.BKDBIN: param.User,
		}
	}

	timeCond := make(map[string]interface{})
	if len(param.Start) > 0 {
		startTime, err := time.Parse(dateLayout, param.Start)
		if err == nil {
			timeCond[pkg.BKDBGTE] = startTime
		}
	}

	if len(param.End) > 0 {
		endTime, err := time.Parse(dateLayout, param.End)
		if err == nil {
			// '%lte: 2006-01-02' means '%lt: 2006-01-03 00:00:00'
			timeCond[pkg.BKDBLT] = endTime.AddDate(0, 0, 1)
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
func (param *GetLaunchHostReq) Validate() error {
	if param.ID <= 0 {
		return errors.New("id should be positive")
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return errors.New("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return errors.New("invalid page.limit < 0")
	}

	if param.Page.Limit > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("exceed page.limit %d", pkg.BKMaxInstanceLimit)
	}

	if param.Filter != nil {
		if _, err := param.Filter.Validate(&querybuilder.RuleOption{NeedSameSliceElementType: true}); err != nil {
			return err
		}
		if param.Filter.GetDeep() > querybuilder.MaxDeep {
			return fmt.Errorf("exceed max query condition deepth: %d", querybuilder.MaxDeep)
		}
	}

	return nil
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
func (param *GetRecallHostReq) Validate() error {
	if param.ID < 0 {
		return errors.New("id can not be negative")
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return errors.New("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return errors.New("invalid page.limit < 0")
	}

	if param.Page.Limit > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("exceed page.limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
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
func (param *GetPoolHostReq) Validate() error {
	for _, selector := range param.Selector {
		if _, err := selector.Validate(); err != nil {
			return err
		}
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return errors.New("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return errors.New("invalid page.limit < 0")
	}

	if param.Page.Limit > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("exceed page.limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
}

// GetFilter get mgo filter
func (param *GetPoolHostReq) GetFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	if len(param.Phase) > 0 {
		filter["status.phase"] = mapstr.MapStr{
			pkg.BKDBIN: param.Phase,
		}
	}

	for _, selector := range param.Selector {
		key := fmt.Sprintf("labels.%s", selector.Key)
		switch selector.Operator {
		case table.SelectOpEqual:
			filter[key] = selector.Value
		case table.SelectOpIn:
			filter[key] = mapstr.MapStr{
				pkg.BKDBIN: selector.Value,
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
func (param *DrawHostReq) Validate() error {
	if len(param.HostIDs) == 0 {
		return errors.New("bk_host_ids cannot be empty")
	}

	if len(param.HostIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("bk_host_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if param.ToBizID <= 0 {
		return errors.New("bk_biz_id should be positive")
	}

	return nil
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
func (param *ReturnHostReq) Validate() error {
	if param.RecallID <= 0 {
		return errors.New("recall_id should be positive")
	}

	if param.FromBizID <= 0 {
		return errors.New("bk_biz_id should be positive")
	}

	if len(param.HostIDs) == 0 {
		return errors.New("bk_host_ids cannot be empty")
	}

	if len(param.HostIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("bk_host_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
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
func (param *CreateRecallOrderReq) Validate() error {
	if param.DeviceType == "" {
		return errors.New("device_type cannot be empty")
	}

	if len(param.AssetIDs) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("asset_ids exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	if param.Replicas <= 0 {
		return errors.New("replicas should be positive")
	}

	if param.Replicas > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("replicas exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
}

// GetRecallOrderReq get resource recall order request
type GetRecallOrderReq struct {
	ID uint64 `json:"id"`
}

// Validate whether GetRecallOrderReq is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetRecallOrderReq) Validate() error {
	if param.ID <= 0 {
		return errors.New("id should be positive")
	}

	return nil
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
func (param *GetRecalledInstReq) Validate() error {
	if param.ID <= 0 {
		return errors.New("id should be positive")
	}

	return nil
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
func (param *GetRecallDetailReq) Validate() error {
	if param.ID <= 0 {
		return errors.New("id should be positive")
	}

	if _, err := param.Page.Validate(false); err != nil {
		return err
	}

	if param.Page.Start < 0 {
		return errors.New("invalid page.start < 0")
	}

	if param.Page.Limit < 0 {
		return errors.New("invalid page.limit < 0")
	}

	if param.Page.Limit > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("exceed page.limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
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
func (param *ResumeRecycleTaskReq) Validate() error {
	if len(param.ID) == 0 {
		return fmt.Errorf("id should be set")
	}

	if len(param.ID) > pkg.BKMaxInstanceLimit {
		return fmt.Errorf("id exceed limit %d", pkg.BKMaxInstanceLimit)
	}

	return nil
}

// GetGradeCfgRst get pool grade config result
type GetGradeCfgRst struct {
	Info []*table.GradeCfg `json:"info"`
}

// GetDeviceTypeRst get pool supported device type result
type GetDeviceTypeRst struct {
	Info []interface{} `json:"info"`
}
