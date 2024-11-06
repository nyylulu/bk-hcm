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

// Package metadata 定义了metadata相关的数据结构
package metadata

import (
	"fmt"
	"strconv"

	"hcm/pkg"
)

// Page for paging query
const (
	PageName         = "page"
	PageSort         = "sort"
	PageStart        = "start"
	PageLimit        = "limit"
	DBFields         = "fields"
	DBQueryCondition = "condition"
)

// BasePage for paging query
type BasePage struct {
	Sort        string `json:"sort,omitempty" mapstructure:"sort"`
	Limit       int    `json:"limit,omitempty" mapstructure:"limit"`
	Start       int    `json:"start" mapstructure:"start"`
	EnableCount bool   `json:"enable_count,omitempty" mapstructure:"enable_count,omitempty"`
}

// Validate page
func (page BasePage) Validate(allowNoLimit bool) (string, error) {
	// 此场景下如果仅仅是获取查询对象的数量，page的其余参数只能是初始化值
	if page.EnableCount {
		if page.Start > 0 || page.Limit > 0 || page.Sort != "" {
			return "page", fmt.Errorf("params page can not be set")
		}
		return "", nil
	}

	if page.Limit > pkg.BKMaxPageSize {
		if page.Limit != pkg.BKNoLimit || allowNoLimit != true {
			return "limit", fmt.Errorf("exceed max page size: %d", pkg.BKMaxPageSize)
		}
	}
	return "", nil
}

// IsIllegal  limit is illegal
func (page BasePage) IsIllegal() bool {
	if page.Limit > pkg.BKMaxPageSize && page.Limit != pkg.BKNoLimit ||
		page.Limit <= 0 {
		return true
	}
	return false
}

// ValidateLimit validates target page limit.
func (page BasePage) ValidateLimit(maxLimit int) error {
	if page.Limit == 0 {
		return fmt.Errorf("page limit must not be zero")
	}

	if maxLimit > pkg.BKMaxPageSize {
		return fmt.Errorf("exceed system max page size: %d", pkg.BKMaxPageSize)
	}

	if page.Limit > maxLimit {
		return fmt.Errorf("exceed business max page size: %d", maxLimit)
	}

	return nil
}

// ParsePage parse page
func ParsePage(origin interface{}) BasePage {
	if origin == nil {
		return BasePage{Limit: pkg.BKNoLimit}
	}
	page, ok := origin.(map[string]interface{})
	if !ok {
		return BasePage{Limit: pkg.BKNoLimit}
	}
	result := BasePage{}
	if sort, ok := page["sort"]; ok && sort != nil {
		result.Sort = fmt.Sprint(sort)
	}
	if start, ok := page["start"]; ok {
		result.Start, _ = strconv.Atoi(fmt.Sprint(start))
	}
	if limit, ok := page["limit"]; ok {
		result.Limit, _ = strconv.Atoi(fmt.Sprint(limit))
		if result.Limit <= 0 {
			result.Limit = pkg.BKNoLimit
		}
	}
	return result
}

// ToSearchSort to search sort
func (page BasePage) ToSearchSort() []SearchSort {
	return NewSearchSortParse().String(page.Sort).ToSearchSortArr()
}
