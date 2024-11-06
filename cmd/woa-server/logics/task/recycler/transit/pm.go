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

// Package transit implements the transit task logic
package transit

import (
	"fmt"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/logs"
)

// TransitPm deals with physical machine transit task
func (t *Transit) TransitPm(order *table.RecycleOrder, hosts []*table.RecycleHost) *event.Event {
	switch order.RecycleType {
	case table.RecycleTypeDissolve, table.RecycleTypeExpired:
		return t.DealTransitTask2Transit(order, hosts)
	case table.RecycleTypeRegular:
		return t.DealTransitTask2Pool(order, hosts)
	default:
		logs.Warnf("failed to deal transit task for order %s, for unknown recycle type %s", order.SuborderID,
			order.RecycleType)
		ev := &event.Event{
			Type: event.TransitFailed,
			Error: fmt.Errorf("failed to deal transit task for order %s, for unknown recycle type %s", order.SuborderID,
				order.RecycleType),
		}
		return ev
	}
}
