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

// Package cmdb ...
package cmdb

const (
	// DftModuleIdle "空闲机"模块
	DftModuleIdle int64 = 1
	// DftModuleFault "故障机"模块
	DftModuleFault int64 = 2
	// DftModuleRecycle "待回收"模块
	DftModuleRecycle int64 = 3
)

const (
	// BusinessSearchMaxLimit 业务搜索接口最大限制
	BusinessSearchMaxLimit int = 200
)
