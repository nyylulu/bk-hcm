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

// Package dao defines all the DAO to be operated.
package dao

// DaoSet defines all the DAO to be operated.
type DaoSet interface {
	ModifyRecord() ModifyRecord
	RecycleOrder() RecycleOrder
	RecycleHost() RecycleHost
	DetectTask() DetectTask
	DetectStep() DetectStep
	DetectStepCfg() DetectStepCfg
	ReturnTask() ReturnTask
	DissolvePlan() DissolvePlan
}

type set struct {
	modifyRecord  ModifyRecord
	recycleOrder  RecycleOrder
	recycleHost   RecycleHost
	detectTask    DetectTask
	detectStep    DetectStep
	detectStepCfg DetectStepCfg
	returnTask    ReturnTask
	dissolvePlan  DissolvePlan
}

var daoSet *set

func init() {
	daoSet = &set{
		modifyRecord:  &modifyRecordDao{},
		recycleOrder:  &recycleOrderDao{},
		recycleHost:   &recycleHostDao{},
		detectTask:    &detectTaskDao{},
		detectStep:    &detectStepDao{},
		detectStepCfg: &detectStepCfgDao{},
		returnTask:    &returnTaskDao{},
		dissolvePlan:  &dissolvePlanDao{},
	}
}

// Set return all dao set interface
func Set() *set {
	return daoSet
}

// ModifyRecord get apply order modify record operation interface
func (s *set) ModifyRecord() ModifyRecord {
	return s.modifyRecord
}

// RecycleOrder get recycle order operation interface
func (s *set) RecycleOrder() RecycleOrder {
	return s.recycleOrder
}

// RecycleHost get recycle host operation interface
func (s *set) RecycleHost() RecycleHost {
	return s.recycleHost
}

// DetectTask get recycle detection task operation interface
func (s *set) DetectTask() DetectTask {
	return s.detectTask
}

// DetectStep get recycle detection step operation interface
func (s *set) DetectStep() DetectStep {
	return s.detectStep
}

// DetectStepCfg get recycle detection step config operation interface
func (s *set) DetectStepCfg() DetectStepCfg {
	return s.detectStepCfg
}

// ReturnTask get recycle return task operation interface
func (s *set) ReturnTask() ReturnTask {
	return s.returnTask
}

// DissolvePlan get resource dissolve plan operation interface
func (s *set) DissolvePlan() DissolvePlan {
	return s.dissolvePlan
}
