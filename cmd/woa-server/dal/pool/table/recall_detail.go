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

// Package table defines the resource recall task's execution detail information
package table

import "time"

// RecallDetail defines a resource recall task's execution detail information
type RecallDetail struct {
	ID             string            `json:"id" bson:"id"`
	RecallID       uint64            `json:"recall_id" bson:"recall_id"`
	HostID         int64             `json:"bk_host_id" bson:"bk_host_id"`
	Labels         map[string]string `json:"labels" bson:"labels"`
	Status         RecallStatus      `json:"status" bson:"status"`
	Message        string            `json:"message" bson:"message"`
	ClearCheckID   string            `json:"clear_check_id" bson:"clear_check_id"`
	ClearCheckLink string            `json:"clear_check_link" bson:"clear_check_link"`
	ReinstallID    string            `json:"reinstall_id" bson:"reinstall_id"`
	ReinstallLink  string            `json:"reinstall_link" bson:"reinstall_link"`
	InitializeID   string            `json:"initialize_id" bson:"initialize_id"`
	InitializeLink string            `json:"initialize_link" bson:"initialize_link"`
	DataDeleteID   string            `json:"data_delete_id" bson:"data_delete_id"`
	DataDeleteLink string            `json:"data_delete_link" bson:"data_delete_link"`
	ConfCheckID    string            `json:"conf_check_id" bson:"conf_check_id"`
	ConfCheckLink  string            `json:"conf_check_link" bson:"conf_check_link"`
	Operator       string            `json:"operator" bson:"operator"`
	CreateAt       time.Time         `json:"create_at" bson:"create_at"`
	UpdateAt       time.Time         `json:"update_at" bson:"update_at"`
}
