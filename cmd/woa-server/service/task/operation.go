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

package task

import (
	"hcm/cmd/woa-server/common"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetApplyStatistics get apply operation statistics
func (s *service) GetApplyStatistics(cts *rest.Contexts) (any, error) {
	input := new(types.GetApplyStatReq)
	if err := cts.DecodeInto(input); err != nil {
		logs.Errorf("failed to get resource apply operation statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	errKey, err := input.Validate()
	if err != nil {
		logs.Errorf("failed to get resource apply operation statistics, err: %v, errKey: %s, rid: %s",
			err, errKey, cts.Kit.Rid)
		return nil, errf.NewFromErr(common.CCErrCommParamsIsInvalid, err)
	}

	rst, err := s.logics.Operation().GetApplyStatistics(cts.Kit, input)
	if err != nil {
		logs.Errorf("failed to get resource apply operation statistics, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return rst, nil
}
