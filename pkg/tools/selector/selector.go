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

// Package selector ...
package selector

import (
	"errors"
	"fmt"

	"hcm/pkg"
	"hcm/pkg/tools/util"
)

// LabelAddOption ...
type LabelAddOption struct {
	InstanceIDs []int64 `json:"instance_ids"`
	Labels      Labels  `json:"labels"`
}

// LabelUpdateOption update instanceIDs labels.
type LabelUpdateOption struct {
	InstanceIDs []int64 `json:"instance_ids"`
	Labels      Labels  `json:"labels"`
}

// LabelUpdateRequest update instanceIDs labels request.
type LabelUpdateRequest struct {
	Option    *LabelUpdateOption `json:"option"`
	TableName string             `json:"table_name"`
}

// LabelAddRequest ...
type LabelAddRequest struct {
	Option    LabelAddOption `json:"option"`
	TableName string         `json:"table_name"`
}

// LabelRemoveOption ...
type LabelRemoveOption struct {
	InstanceIDs []int64  `json:"instance_ids"`
	Keys        []string `json:"keys"`
}

// LabelRemoveRequest ...
type LabelRemoveRequest struct {
	Option    LabelRemoveOption `json:"option"`
	TableName string            `json:"table_name"`
}

// Operator ...
type Operator string

const (
	// DoesNotExist ...
	DoesNotExist Operator = "!"
	// Equals ...
	Equals Operator = "="
	// In ...
	In Operator = "in"
	// NotEquals ...
	NotEquals Operator = "!="
	// NotIn ...
	NotIn Operator = "notin"
	// Exists ...
	Exists Operator = "exists"
)

// AvailableOperators ...
var AvailableOperators = []Operator{
	DoesNotExist,
	Equals,
	In,
	NotEquals,
	NotIn,
	Exists,
}

// Selector ...
type Selector struct {
	Key      string   `json:"key" field:"key" bson:"key"`
	Operator Operator `json:"operator" field:"operator" bson:"operator"`
	Values   []string `json:"values" field:"values" bson:"values"`
}

// Validate ...
func (s *Selector) Validate() (string, error) {
	if util.InArray(s.Operator, AvailableOperators) == false {
		return "operator", fmt.Errorf("operator %s not available, available operators: %+v", s.Operator,
			AvailableOperators)
	}

	if (s.Operator == In || s.Operator == NotIn) && len(s.Values) == 0 {
		return "values", errors.New("values shouldn't be empty")
	}

	if (s.Operator == Exists || s.Operator == DoesNotExist) && len(s.Values) > 0 {
		return "values", errors.New("values shouldn be empty")
	}

	if (s.Operator == Equals || s.Operator == NotEquals) && len(s.Values) != 1 {
		return "values", errors.New("values field length for equal operation should exactly one")
	}

	if LabelNGKeyRule.MatchString(s.Key) == false {
		return "key", fmt.Errorf("key %s invalid", s.Key)
	}
	return "", nil
}

// ToMgoFilter ...
func (s *Selector) ToMgoFilter() (map[string]interface{}, error) {
	filter := make(map[string]interface{})
	field := "labels." + s.Key
	switch s.Operator {
	case In:
		filter = map[string]interface{}{
			field: map[string]interface{}{
				pkg.BKDBIN: s.Values,
			},
		}
	case NotIn:
		filter = map[string]interface{}{
			field: map[string]interface{}{
				pkg.BKDBNIN: s.Values,
			},
		}
	case DoesNotExist, Exists:
		filter = map[string]interface{}{
			field: map[string]interface{}{
				pkg.BKDBExists: s.Operator == Exists,
			},
		}
	case Equals:
		if len(s.Values) == 0 {
			return nil, errors.New("values empty")
		}
		firstValue := s.Values[0]
		filter = map[string]interface{}{
			field: firstValue,
		}
	case NotEquals:
		if len(s.Values) == 0 {
			return nil, errors.New("values empty")
		}
		firstValue := s.Values[0]
		filter = map[string]interface{}{
			field: map[string]interface{}{
				pkg.BKDBNE: firstValue,
			},
		}
	}
	return filter, nil
}

// Selectors ...
type Selectors []Selector

// Validate ...
func (ss Selectors) Validate() (string, error) {
	for _, selector := range ss {
		if key, err := selector.Validate(); err != nil {
			return key, err
		}
	}
	return "", nil
}

// ToMgoFilter ...
func (ss Selectors) ToMgoFilter() (map[string]interface{}, error) {
	filters := make([]map[string]interface{}, 0)
	for _, selector := range ss {
		filter, err := selector.ToMgoFilter()
		if err != nil {
			return nil, err
		}
		filters = append(filters, filter)
	}
	if len(filters) == 0 {
		return make(map[string]interface{}), nil
	}
	if len(filters) == 1 {
		return filters[0], nil
	}
	return map[string]interface{}{
		pkg.BKDBAND: filters,
	}, nil
}
