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

// Package task ...
package task

// CVM create cvm request param
type CVM struct {
	AppId             string `json:"appId"`
	ApplyType         int64  `json:"applyType"`
	AppModuleId       int64  `json:"appModuleId"`
	Operator          string `json:"operator"`
	ApplyNumber       uint   `json:"applyNumber"`
	NoteInfo          string `json:"noteInfo"`
	VPCId             string `json:"vpcId"`
	SubnetId          string `json:"subnetId"`
	Area              string `json:"area"`
	Zone              string `json:"zone"`
	ImageId           string `json:"image_id"`
	ImageName         string `json:"image_name"`
	InstanceType      string `json:"instanceType"`
	DiskType          string `json:"disk_type"`
	DiskSize          int64  `json:"disk_size"`
	SecurityGroupId   string `json:"securityGroupId"`
	SecurityGroupName string `json:"securityGroupName"`
	SecurityGroupDesc string `json:"securityGroupDesc"`
}
