/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017,-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package metadata ...
package metadata

import (
	"time"

	"hcm/cmd/woa-server/common"
)

// ChartConfig ...
type ChartConfig struct {
	ConfigID   uint64 `json:"config_id" bson:"config_id"`
	ReportType string `json:"report_type" bson:"report_type"`
	Name       string `json:"name" bson:"name"`
	CreateTime Time   `json:"create_time" bson:"create_time"`
	OwnerID    string `json:"bk_supplier_account" bson:"bk_supplier_account"`
	ObjID      string `json:"bk_obj_id" bson:"bk_obj_id"`
	Width      string `json:"width" bson:"width"`
	ChartType  string `json:"chart_type" bson:"chart_type"`
	Field      string `json:"field" bson:"field"`
	XAxisCount int64  `json:"x_axis_count" bson:"x_axis_count"`
}

// ChartPosition chart position
type ChartPosition struct {
	BizID    int64        `json:"bk_biz_id" bson:"bk_biz_id"`
	Position PositionInfo `json:"position" bson:"position"`
	OwnerID  string       `json:"bk_supplier_account" bson:"bk_supplier_account"`
}

// PositionInfo chart position
type PositionInfo struct {
	Host []uint64 `json:"host" bson:"host"`
	Inst []uint64 `json:"inst" bson:"inst"`
}

// ModelInstChange model inst change
type ModelInstChange map[string]*InstChangeCount

// InstChangeCount inst change count
type InstChangeCount struct {
	Create int64 `json:"create" bson:"create"`
	Update int64 `json:"update" bson:"update"`
	Delete int64 `json:"delete" bson:"delete"`
}

// AggregateIntResponse aggregate int response
type AggregateIntResponse struct {
	BaseResp `json:",inline"`
	Data     []IntIDCount `json:"data"`
}

// IntIDCount int类型字段做mongoDB聚合时使用
type IntIDCount struct {
	ID    int64 `json:"id" bson:"_id"`
	Count int64 `json:"count" bson:"count"`
}

// IntIDArrayCount int类型字段做mongoDB聚合，且结果为数组时使用
type IntIDArrayCount struct {
	ID    int64   `json:"id" bson:"_id"`
	Count []int64 `json:"count" bson:"count"`
}

// AggregateStringResponse aggregate string response
type AggregateStringResponse struct {
	BaseResp `json:",inline"`
	Data     []StringIDCount `json:"data"`
}

// StringIDCount string类型字段做mongoDB聚合时使用
type StringIDCount struct {
	ID    string `json:"id" bson:"_id"`
	Count int64  `json:"count" bson:"count"`
}

// ObjectIDCount object count statistics information used for
// group aggregate operation.
type ObjectIDCount struct {
	// ObjID object id.
	ObjID string `bson:"_id" json:"bk_obj_id"`

	// Count targets count.
	Count int64 `bson:"count" json:"instance_count"`
}

// UpdateInstCount update instance count.
type UpdateInstCount struct {
	ID    UpdateID `json:"id" bson:"_id"`
	Count int64    `json:"count" bson:"count"`
}

// UpdateID update instance id.
type UpdateID struct {
	ObjID  string `json:"bk_obj_id" bson:"bk_obj_id"`
	InstID int64  `json:"bk_inst_id" bson:"bk_inst_id"`
}

// HostChangeChartData host change chart data
type HostChangeChartData struct {
	ReportType string          `json:"report_type" bson:"report_type"`
	Data       []StringIDCount `json:"data" bson:"data"`
	OwnerID    string          `json:"bk_supplier_account" bson:"bk_supplier_account"`
	CreateTime string          `json:"create_time" bson:"create_time"`
}

// ChartData chart data
type ChartData struct {
	ReportType string      `json:"report_type" bson:"report_type"`
	Data       interface{} `json:"data" data:"data"`
	OwnerID    string      `json:"bk_supplier_account" bson:"bk_supplier_account"`
	LastTime   time.Time   `json:"last_time" bson:"last_time"`
}

// ModelInstChartData model instance chart data
type ModelInstChartData struct {
	ReportType string          `json:"report_type" bson:"report_type"`
	Data       []StringIDCount `json:"data" data:"data"`
	OwnerID    string          `json:"bk_supplier_account" bson:"bk_supplier_account"`
	LastTime   time.Time       `json:"last_time" bson:"last_time"`
}

// SearchChartResponse search chart response
type SearchChartResponse struct {
	BaseResp `json:",inline"`
	Data     SearchChartConfig `json:"data"`
}

// SearchChartCommon search chart common
type SearchChartCommon struct {
	BaseResp `json:",inline"`
	Data     CommonSearchChart `json:"data"`
}

// CommonSearchChart common search chart
type CommonSearchChart struct {
	Count uint64      `json:"count"`
	Info  ChartConfig `json:"info"`
}

// SearchChartConfig search chart config
type SearchChartConfig struct {
	Count uint64                   `json:"count"`
	Info  map[string][]ChartConfig `json:"info"`
}

// CloudMapping cloud mapping
type CloudMapping struct {
	CreateTime Time   `json:"create_time" bson:"create_time"`
	LastTime   Time   `json:"last_time" bson:"lsat_time"`
	CloudName  string `json:"bk_cloud_name" bson:"bk_cloud_name"`
	OwnerID    string `json:"bk_supplier_account" bson:"bk_supplier_account"`
	CloudID    int64  `json:"bk_cloud_id" bson:"bk_cloud_id"`
}

// ChartClassification chart classification
type ChartClassification struct {
	Host []ChartConfig `json:"host"`
	Inst []ChartConfig `json:"inst"`
	Nav  []ChartConfig `json:"nav"`
}

// ObjectIDName object id name
type ObjectIDName struct {
	ObjectID   string `json:"bk_object_id"`
	ObjectName string `json:"bk_object_name"`
}

// StatisticInstOperation statistic inst operation
type StatisticInstOperation struct {
	Create []StringIDCount   `json:"create"`
	Delete []StringIDCount   `json:"delete"`
	Update []UpdateInstCount `json:"update"`
}

var (
	// BizModuleHostChart biz module host chart
	BizModuleHostChart = ChartConfig{
		ReportType: common.BizModuleHostChart,
	}

	// HostOsChart host os chart
	HostOsChart = ChartConfig{
		ReportType: common.HostOSChart,
		Name:       "按操作系统类型统计",
		ObjID:      "host",
		Width:      "50",
		ChartType:  "pie",
		Field:      "bk_os_type",
		XAxisCount: 10,
	}

	// HostBizChart host biz chart
	HostBizChart = ChartConfig{
		ReportType: common.HostBizChart,
		Name:       "按业务统计",
		ObjID:      "host",
		Width:      "50",
		ChartType:  "bar",
		XAxisCount: 10,
	}

	// HostCloudChart host cloud chart
	HostCloudChart = ChartConfig{
		ReportType: common.HostCloudChart,
		Name:       "按云区域统计",
		Width:      "100",
		ObjID:      "host",
		ChartType:  "bar",
		Field:      common.BKCloudIDField,
		XAxisCount: 20,
	}

	// HostChangeBizChart host change biz chart
	HostChangeBizChart = ChartConfig{
		ReportType: common.HostChangeBizChart,
		Name:       "主机数量变化趋势",
		Width:      "100",
		XAxisCount: 20,
	}

	// ModelAndInstCountChart model and inst count chart
	ModelAndInstCountChart = ChartConfig{
		ReportType: common.ModelAndInstCount,
	}

	// ModelInstChart model inst chart
	ModelInstChart = ChartConfig{
		ReportType: common.ModelInstChart,
		Name:       "实例数量统计",
		Width:      "50",
		ChartType:  "bar",
		XAxisCount: 10,
	}

	// ModelInstChangeChart model inst change chart
	ModelInstChangeChart = ChartConfig{
		ReportType: common.ModelInstChangeChart,
		Name:       "实例变更统计",
		Width:      "50",
		ChartType:  "bar",
		XAxisCount: 10,
	}

	// InnerChartsMap inner charts map
	InnerChartsMap = map[string]ChartConfig{
		common.BizModuleHostChart:   BizModuleHostChart,
		common.ModelAndInstCount:    ModelAndInstCountChart,
		common.HostOSChart:          HostOsChart,
		common.HostBizChart:         HostBizChart,
		common.HostCloudChart:       HostCloudChart,
		common.HostChangeBizChart:   HostChangeBizChart,
		common.ModelInstChart:       ModelInstChart,
		common.ModelInstChangeChart: ModelInstChangeChart,
	}

	// InnerChartsArr inner charts array
	InnerChartsArr = []string{
		common.BizModuleHostChart,
		common.ModelAndInstCount,
		common.HostOSChart,
		common.HostBizChart,
		common.HostCloudChart,
		common.HostChangeBizChart,
		common.ModelInstChart,
		common.ModelInstChangeChart,
	}
)
