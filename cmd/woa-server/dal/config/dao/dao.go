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

// Package dao implements data access object which provides an abstract interface to database
package dao

// DaoSet defines all the DAO to be operated.
type DaoSet interface {
	ZoneLeftIP() ZoneLeftIP
}

type set struct {
	zoneLeftIP ZoneLeftIP
}

var daoSet *set

func init() {
	daoSet = &set{
		zoneLeftIP: &zoneLeftIPDao{},
	}
}

// Set return all dao set interface
func Set() *set {
	return daoSet
}

// ZoneLeftIP get zone with left ip operation interface
func (s *set) ZoneLeftIP() ZoneLeftIP {
	return s.zoneLeftIP
}
