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

// Package recommender provides the ability to automatically analyze various available resource
// and make recommendations for resource apply
package recommender

import (
	"context"

	"hcm/cmd/woa-server/common/mapstr"
	"hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty"
	"hcm/pkg/thirdparty/cvmapi"
)

// Recommender make recommendations for resource apply
type Recommender struct {
	cvm cvmapi.CVMClientInterface
	ctx context.Context

	handlers []Handler
}

// New create a recommender
func New(ctx context.Context, thirdCli *thirdparty.Client) (*Recommender, error) {
	handlers := initRecommendHandlers(thirdCli.CVM)

	recommend := &Recommender{
		cvm:      thirdCli.CVM,
		handlers: handlers,
		ctx:      ctx,
	}

	return recommend, nil
}

// GetApplyRecommendation get result of apply order modification recommendation
func (r *Recommender) GetApplyRecommendation(suborderID string) (*types.RecommendApplyRst, error) {
	filter := &mapstr.MapStr{
		"suborder_id": suborderID,
	}

	order, err := model.Operation().ApplyOrder().GetApplyOrder(context.Background(), filter)
	if err != nil {
		logs.Errorf("failed to get apply order, err: %v", err)
		return nil, err
	}

	recommend := GetRecommendationByChain(order, r.handlers)

	rst := &types.RecommendApplyRst{
		SuborderID: order.SubOrderId,
		Replicas:   order.Total,
		Spec:       order.Spec,
	}

	rst.Spec.Zone = recommend.Zone
	rst.Spec.DeviceType = recommend.DeviceType

	return rst, nil
}
