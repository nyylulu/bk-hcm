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

// Package table define the table structure of the resource pool
package table

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"hcm/pkg/tools/util"

	jsoniter "github.com/json-iterator/go"
)

// OpTaskPhase resource operation task phase
type OpTaskPhase string

// definition of various resource operation task phase
const (
	OpTaskPhaseInit    OpTaskPhase = "INIT"
	OpTaskPhaseRunning OpTaskPhase = "RUNNING"
	OpTaskPhasePaused  OpTaskPhase = "PAUSED"
	OpTaskPhaseSuccess OpTaskPhase = "SUCCESS"
	OpTaskPhaseFailed  OpTaskPhase = "FAILED"
)

// definition of various resource operation task phase description
const (
	OpTaskPhaseDescInit    string = "未执行"
	OpTaskPhaseDescRunning string = "执行中"
	OpTaskPhaseDescPaused  string = "已暂停"
	OpTaskPhaseDescSuccess string = "成功"
	OpTaskPhaseDescFailed  string = "失败"
)

// ResourceType resource type
type ResourceType string

// definition of various resource type
const (
	ResourceTypePm          ResourceType = "IDCPM"
	ResourceTypeCvm         ResourceType = "QCLOUDCVM"
	ResourceTypeIdcDvm      ResourceType = "IDCDVM"
	ResourceTypeQcloudDvm   ResourceType = "QCLOUDDVM"
	ResourceTypeOthers      ResourceType = "OTHERS"
	ResourceTypeUnsupported ResourceType = "UNSUPPORTED"
)

// OpType resource operation type
type OpType string

// definition of various resource operation type
const (
	OpTypeLaunch  OpType = "LAUNCH"
	OpTypeRecall  OpType = "RECALL"
	OpTypeRecycle OpType = "RECYCLE"
)

// PoolHostPhase resource pool host phase
type PoolHostPhase string

// definition of various resource pool host phases
const (
	PoolHostPhaseLaunching PoolHostPhase = "LAUNCHING"
	PoolHostPhaseIdle      PoolHostPhase = "IDLE"
	PoolHostPhaseInUse     PoolHostPhase = "IN_USE"
	PoolHostPhaseForRecall PoolHostPhase = "FOR_RECALL"
	PoolHostPhaseRecalled  PoolHostPhase = "RECALLED"
)

// Tag object tag
type Tag struct {
	Key   string `json:"key" bson:"key"`
	Value string `json:"value" bson:"value"`
}

// Selector label selector
type Selector struct {
	Key      string      `json:"key" bson:"key"`
	Operator SelectOp    `json:"op" bson:"op"`
	Value    interface{} `json:"value" bson:"value"`
}

// Validate validates Selector parameters
func (s Selector) Validate() (string, error) {
	// 此场景下如果仅仅是获取查询对象的数量，page的其余参数只能是初始化值
	if len(s.Key) == 0 {
		return "key", errors.New("key can not be empty")
	}

	switch s.Operator {
	case SelectOpEqual, SelectOpNotEqual:
		if err := s.validateBasicType(s.Value); err != nil {
			return "value", err
		}
	case SelectOpIn, SelectOpNotIn:
		if err := s.validateSliceOfBasicType(s.Value, true, 0); err != nil {
			return "value", err
		}
	default:
		return "op", fmt.Errorf("unsupported op: %s", s.Operator)
	}

	return "", nil
}

var (
	TypeNumeric = "numeric"
	TypeBoolean = "boolean"
	TypeString  = "string"
	TypeUnknown = "unknown"
)

func getType(value interface{}) string {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float64, float32,
		jsoniter.Number, json.Number:
		return TypeNumeric
	case bool:
		return TypeBoolean
	case string:
		return TypeString
	default:
		return TypeUnknown
	}
}

func (s Selector) validateBasicType(value interface{}) error {
	if t := getType(value); t == TypeUnknown {
		return fmt.Errorf("unknow value type: %v with value: %+v", reflect.TypeOf(value), value)
	}
	return nil
}

func (s Selector) validateSliceOfBasicType(value interface{}, requireSameType bool, maxElementsCount int) error {
	if value == nil {
		return nil
	}

	t := reflect.TypeOf(value)
	if t.Kind() != reflect.Array && t.Kind() != reflect.Slice {
		return fmt.Errorf("unexpected value type: %s, expect array", t.Kind().String())
	}

	v := reflect.ValueOf(value)
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		if err := s.validateBasicType(item); err != nil {
			return err
		}
	}

	if maxElementsCount > 0 && v.Len() > maxElementsCount {
		return fmt.Errorf("too many elements of slice: %d max(%d)", v.Len(), maxElementsCount)
	}

	if requireSameType {
		vTypes := make([]string, 0)
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			vTypes = append(vTypes, getType(item))
		}
		vTypes = util.StrArrayUnique(vTypes)
		if len(vTypes) > 1 {
			return fmt.Errorf("slice element type not unique, types: %+v", vTypes)
		}
	}

	return nil
}

// SelectOp select operation type
type SelectOp string

// definition of various select operation types
const (
	SelectOpEqual    SelectOp = "equal"
	SelectOpNotEqual SelectOp = "not_equal"
	SelectOpIn       SelectOp = "in"
	SelectOpNotIn    SelectOp = "not_in"
)

// RecallStatus resource recall status
type RecallStatus string

// definition of various recall status
const (
	RecallStatusDefault          RecallStatus = "DEFAULT"
	RecallStatusReturned         RecallStatus = "RETURNED"
	RecallStatusPreChecking      RecallStatus = "PRE_CHECKING"
	RecallStatusPreCheckFailed   RecallStatus = "PRE_CHECK_FAILED"
	RecallStatusClearChecking    RecallStatus = "CLEAR_CHECKING"
	RecallStatusClearCheckFailed RecallStatus = "CLEAR_CHECK_FAILED"
	RecallStatusReinstalling     RecallStatus = "REINSTALLING"
	RecallStatusReinstallFailed  RecallStatus = "REINSTALL_FAILED"
	RecallStatusInitializing     RecallStatus = "INITIALIZING"
	RecallStatusInitializeFailed RecallStatus = "INITIALIZE_FAILED"
	RecallStatusDataDeleting     RecallStatus = "DATA_DELETING"
	RecallStatusDataDeleteFailed RecallStatus = "DATA_DELETE_FAILED"
	RecallStatusConfChecking     RecallStatus = "CONF_CHECKING"
	RecallStatusConfCheckFailed  RecallStatus = "CONF_CHECK_FAILED"
	RecallStatusTransiting       RecallStatus = "TRANSITING"
	RecallStatusTransitFailed    RecallStatus = "TRANSIT_FAILED"
	RecallStatusDone             RecallStatus = "DONE"
	RecallStatusTerminate        RecallStatus = "TERMINATE"
)

// definition of various recall status description
const (
	RecallStatusDescReturned         string = "已归还"
	RecallStatusDescPreChecking      string = "准入检查中"
	RecallStatusDescPreCheckFailed   string = "准入检查失败"
	RecallStatusDescClearChecking    string = "空闲检查中"
	RecallStatusDescClearCheckFailed string = "空闲检查失败"
	RecallStatusDescReinstalling     string = "系统重装中"
	RecallStatusDescReinstallFailed  string = "系统重装失败"
	RecallStatusDescInitializing     string = "初始化中"
	RecallStatusDescInitializeFailed string = "初始化失败"
	RecallStatusDescDataDeleting     string = "数据清理中"
	RecallStatusDescDataDeleteFailed string = "数据清理失败"
	RecallStatusDescConfChecking     string = "配置检查中"
	RecallStatusDescConfCheckFailed  string = "配置检查失败"
	RecallStatusDescTransiting       string = "转模块中"
	RecallStatusDescTransitFailed    string = "转模块失败"
	RecallStatusDescDone             string = "已完成"
	RecallStatusDescTerminate        string = "终止"
)
