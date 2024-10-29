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

// Package util ...
package util

import (
	"hcm/pkg"
)

// SetQueryOwner returns condition that in default ownerID and request ownerID
func SetQueryOwner(condition map[string]interface{}, ownerID string) map[string]interface{} {
	if nil == condition {
		condition = make(map[string]interface{})
	}
	if ownerID == pkg.BKSuperOwnerID {
		return condition
	}
	if ownerID == pkg.BKDefaultOwnerID {
		condition[pkg.BKOwnerIDField] = pkg.BKDefaultOwnerID
		return condition
	}
	condition[pkg.BKOwnerIDField] = map[string]interface{}{pkg.BKDBIN: []string{pkg.BKDefaultOwnerID, ownerID}}
	return condition
}

// SetModOwner set condition equal owner id, the condition must be a map or struct
func SetModOwner(condition map[string]interface{}, ownerID string) map[string]interface{} {
	if nil == condition {
		condition = make(map[string]interface{})
	}
	if ownerID == pkg.BKSuperOwnerID {
		return condition
	}
	condition[pkg.BKOwnerIDField] = ownerID
	return condition
}
