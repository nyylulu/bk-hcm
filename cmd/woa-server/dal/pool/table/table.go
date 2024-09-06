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

// Package table provides database table interface
package table

import "fmt"

// Tables defines all the database table
// related resources.
type Tables interface {
	TableName() Name
}

// Name is database table's name type
type Name string

// Validate whether the table name is valid or not.
func (n Name) Validate() error {
	switch n {
	case PoolHostTable:
	case LaunchTaskTable:
	case RecallTaskTable:
	case RecallOrderTable:
	case RecycleTaskTable:
	case RecallDetailTable:
	case PoolOpRecordTable:
	case PoolGradeCfgTable:
	default:
		return fmt.Errorf("unknown table name: %s", n)
	}

	return nil
}

const (
	// PoolHostTable the table name of resource pool host
	PoolHostTable = "cr_PoolHost"
	// LaunchTaskTable the table name of resource launch task
	LaunchTaskTable = "cr_PoolLaunchTask"
	// RecallTaskTable the table name of resource recall task
	RecallTaskTable = "cr_PoolRecallTask"
	// RecallOrderTable the table name of resource recall order
	RecallOrderTable = "cr_PoolRecallOrder"
	// RecallDetailTable the table name of resource recall task detail
	RecallDetailTable = "cr_PoolRecallDetail"
	// RecycleTaskTable the table name of resource recycle task
	RecycleTaskTable = "cr_PoolRecycleTask"
	// PoolOpRecordTable the table name of resource pool operation record
	PoolOpRecordTable = "cr_PoolOpRecord"
	// PoolGradeCfgTable the table name of resource pool grade config
	PoolGradeCfgTable = "cr_CfgPoolGrade"
)
