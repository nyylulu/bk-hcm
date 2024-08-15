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

// Package apply define the key of apply order informer
package apply

import "hcm/cmd/woa-server/common"

// KeyApply apply order informer key for token handler
var KeyApply = Key{
	namespace:  "apply_order",
	collection: common.BKTableNameApplyOrder,
	ttlSeconds: 6 * 60 * 60,
}

// Key key for token handler
type Key struct {
	namespace string
	// the watching db collection name
	collection string
	// the valid event's life time.
	// if the event is exist longer than this, it will be deleted.
	// if use's watch start from value is older than time.Now().Unix() - startFrom value,
	// that means use's is watching event that has already deleted, it's not allowed.
	ttlSeconds int64

	// validator validate whether the event data is valid or not.
	// if not, then this event should not be handle, should be dropped.
	validator func(doc []byte) error

	// instance name returns a name which can describe the event's instances
	instName func(doc []byte) string

	// instID returns the event's corresponding instance id,
	instID func(doc []byte) int64
}

// DetailKey get key's detail key
func (k Key) DetailKey(cursor string) string {
	return k.namespace + ":detail:" + cursor
}

// Namespace get key's namespace
func (k Key) Namespace() string {
	return k.namespace
}

// TTLSeconds get key's ttl in seconds
func (k Key) TTLSeconds() int64 {
	return k.ttlSeconds
}

// Validate validate the key
func (k Key) Validate(doc []byte) error {
	if k.validator != nil {
		return k.validator(doc)
	}

	return nil
}

// Name get key's instance name
func (k Key) Name(doc []byte) string {
	if k.instName != nil {
		return k.instName(doc)
	}
	return ""
}

// InstanceID get key's instance id
func (k Key) InstanceID(doc []byte) int64 {
	if k.instID != nil {
		return k.instID(doc)
	}
	return 0
}

// Collection get key's collection
func (k Key) Collection() string {
	return k.collection
}

// ChainCollection get the event chain db collection name
func (k Key) ChainCollection() string {
	return k.collection + "WatchChain"
}
