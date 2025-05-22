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

// Package pool pool server types
package pool

const (

	/* 基本业务逻辑
	上架  匹配业务/匹配模块 (BizIDMatch/ModuleIDPoolMatch) -> 资源池业务/资源池模块 (BizIDPool/ModuleIDPool)
	借出（机器提取）资源池业务/资源池模块 (BizIDPool/ModuleIDPool) -> 用户业务-用户模块
	归还：用户业务/用户模块 -> 资源池业务/下架中转模块 (BizIDPool/ModuleIDPoolRecalling)
	回收：资源池业务/下架中转模块 (BizIDPool/ModuleIDPoolRecalling) -> 匹配业务/匹配模块 (BizIDMatch/ModuleIDPoolMatch)

	具体：
	上架： 资源运营服务-SA云化池 -> 资源运营服务-CR资源池
	借出： 资源运营服务-CR资源池 -> 用户业务-用户模块
	归还： 用户业务-用户模块 -> 资源运营服务-CR资源下架中
	回收： 资源运营服务-CR资源下架中 -> 资源运营服务-SA云化池
	*/

	// BizIDMatch 资源上架源业务
	BizIDMatch int64 = 931
	// BizIDPool biz id of 资源运营服务
	BizIDPool int64 = 931
	// ModuleIDPool module id of CR资源池
	ModuleIDPool int64 = 5077039
	// ModuleIDPoolRecalling module id of CR资源下架中
	ModuleIDPoolRecalling int64 = 5085334
	// ModuleIDPoolRecallFailed module id of CR资源下架失败
	ModuleIDPoolRecallFailed int64 = 5008422
	// ModuleIDPoolMatch module id of SA云化池
	ModuleIDPoolMatch int64 = 239149
)
