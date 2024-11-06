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

// Package algorithm provides ...
package algorithm

import (
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/thirdparty/dvmapi"
)

// FitPredicate host filter functor
type FitPredicate func(selector *types.DVMSelector, host *dvmapi.DockerHost) (bool, error)

// PriorityFunction host priority functor
type PriorityFunction func(selector *types.DVMSelector, hosts []*dvmapi.DockerHost) (types.HostPriorityList, error)

// PriorityConfig host priority config
type PriorityConfig struct {
	Function PriorityFunction
	Weight   int
}
