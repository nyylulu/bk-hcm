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

// Package recommender apply order modification recommend handler
package recommender

import (
	"hcm/cmd/woa-server/thirdparty/cvmapi"
	types "hcm/cmd/woa-server/types/task"
)

// Handler apply order modification recommend handler
type Handler interface {
	Handle(order *types.ApplyOrder) (*Recommendation, bool)
}

var handlers []Handler

func initRecommendHandlers(cvm cvmapi.CVMClientInterface) []Handler {
	deviceTypeHandler := &DeviceTypeHandler{
		cvm: cvm,
	}
	zoneHandler := &ZoneHandler{
		cvm: cvm,
	}

	handlers = []Handler{
		deviceTypeHandler,
		zoneHandler,
	}

	return handlers
}

// Recommendation apply order modification recommendation
type Recommendation struct {
	Zone       string `json:"zone"`
	DeviceType string `json:"device_type"`
}

// GetRecommendationByChain get result of apply order modification recommendation
// which processed by chain of recommend handlers
func GetRecommendationByChain(order *types.ApplyOrder, handlers []Handler) *Recommendation {
	for _, handler := range handlers {
		rst, ok := handler.Handle(order)
		if ok {
			return rst
		}
	}

	dftRst := &Recommendation{
		Zone:       order.Spec.Zone,
		DeviceType: order.Spec.DeviceType,
	}

	return dftRst
}
