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

package config

import (
	"errors"
	"fmt"

	"hcm/cmd/woa-server/common/metadata"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/cmd/woa-server/dal/config/table"
)

// GetLeftIPParam get zone with left ip parameter
type GetLeftIPParam struct {
	Filter *querybuilder.QueryFilter `json:"filter" bson:"filter"`
	Page   metadata.BasePage         `json:"page" bson:"page"`
}

// Validate whether GetLeftIPParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *GetLeftIPParam) Validate() (errKey string, err error) {
	if key, err := param.Page.Validate(true); err != nil {
		return fmt.Sprintf("page.%s", key), err
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
func (param *GetLeftIPParam) GetFilter() (map[string]interface{}, error) {
	if param.Filter != nil {
		mgoFilter, key, err := param.Filter.ToMgo()
		if err != nil {
			return nil, fmt.Errorf("invalid key:filter.%s, err: %s", key, err)
		}
		return mgoFilter, nil
	}
	return make(map[string]interface{}), nil
}

// GetLeftIPRst get zone with left ip result
type GetLeftIPRst struct {
	Count int64               `json:"count"`
	Info  []*table.ZoneLeftIP `json:"info"`
}

// UpdateLeftIPPropertyParam update zone with left ip property request param
type UpdateLeftIPPropertyParam struct {
	Ids      []int64                `json:"ids" bson:"ids"`
	Property map[string]interface{} `json:"properties"`
}

// Validate whether UpdateLeftIPPropertyParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *UpdateLeftIPPropertyParam) Validate() (errKey string, err error) {
	if len(param.Ids) <= 0 {
		return "ids", errors.New("cannot be empty")
	}

	limit := 200
	if len(param.Ids) > limit {
		return "ids", fmt.Errorf("exceed limit %d", limit)
	}

	return "", nil
}

// SyncLeftIPParam sync zone left IP request param
type SyncLeftIPParam struct {
	Region string `json:"region"`
	Zone   string `json:"zone"`
}

// Validate whether SyncLeftIPParam is valid
// errKey: invalid key
// err: detail reason why errKey is invalid
func (param *SyncLeftIPParam) Validate() (errKey string, err error) {
	if len(param.Region) <= 0 {
		return "region", errors.New("cannot be empty")
	}

	if len(param.Zone) <= 0 {
		return "zone", errors.New("cannot be empty")
	}

	return "", nil
}
