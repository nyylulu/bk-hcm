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

// Package metadata define the metadata struct
package metadata

import (
	"fmt"
	"sort"
	"strings"

	"hcm/cmd/woa-server/common/mapstr"
)

// ObjectUnique define the unique attribute of object
type ObjectUnique struct {
	ID       uint64      `json:"id" bson:"id"`
	ObjID    string      `json:"bk_obj_id" bson:"bk_obj_id"`
	Keys     []UniqueKey `json:"keys" bson:"keys"`
	Ispre    bool        `json:"ispre" bson:"ispre"`
	OwnerID  string      `json:"bk_supplier_account" bson:"bk_supplier_account"`
	LastTime Time        `json:"last_time" bson:"last_time"`
}

// Parse load the data from mapstr attribute into ObjectUnique instance
func (cli *ObjectUnique) Parse(data mapstr.MapStr) (*ObjectUnique, error) {

	err := mapstr.SetValueToStructByTags(cli, data)
	if nil != err {
		return nil, err
	}

	return cli, err
}

// KeysHash return the unique keys hash
func (u ObjectUnique) KeysHash() string {
	keys := []string{}
	for _, key := range u.Keys {
		keys = append(keys, fmt.Sprintf("%s:%d", key.Kind, key.ID))
	}
	sort.Strings(keys)
	return strings.Join(keys, "#")
}

// UniqueKey define the unique key
type UniqueKey struct {
	Kind string `json:"key_kind" bson:"key_kind"`
	ID   uint64 `json:"key_id" bson:"key_id"`
}

const (
	// UniqueKeyKindProperty property
	UniqueKeyKindProperty = "property"
	// UniqueKeyKindAssociation association
	UniqueKeyKindAssociation = "association"
)

// CreateUniqueRequest create unique request
type CreateUniqueRequest struct {
	ObjID string      `json:"bk_obj_id" bson:"bk_obj_id"`
	Keys  []UniqueKey `json:"keys" bson:"keys"`
}

// RspID response id
type RspID struct {
	ID int64 `json:"id"`
}

// CreateUniqueResult create unique result
type CreateUniqueResult struct {
	BaseResp
	Data RspID `json:"data"`
}

// UpdateUniqueRequest update unique request
type UpdateUniqueRequest struct {
	Keys     []UniqueKey `json:"keys" bson:"keys"`
	LastTime Time        `json:"last_time" bson:"last_time"`
}

// UpdateUniqueResult update unique result
type UpdateUniqueResult struct {
	BaseResp
}

// DeleteUniqueRequest delete unique request
type DeleteUniqueRequest struct {
	ID    uint64 `json:"id"`
	ObjID string `json:"bk_obj_id"`
}

// DeleteUniqueResult delete unique result
type DeleteUniqueResult struct {
	BaseResp
}

// SearchUniqueRequest search unique request
type SearchUniqueRequest struct {
	ObjID string `json:"bk_obj_id"`
}

// SearchUniqueResult search unique result
type SearchUniqueResult struct {
	BaseResp
	Data []ObjectUnique `json:"data"`
}

// QueryUniqueResult query unique result
type QueryUniqueResult struct {
	Count uint64         `json:"count"`
	Info  []ObjectUnique `json:"info"`
}
