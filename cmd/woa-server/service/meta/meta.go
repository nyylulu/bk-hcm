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

package meta

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/types/meta"
	"hcm/pkg/rest"
)

// ListDiskType lists disk type.
func (s *service) ListDiskType(_ *rest.Contexts) (interface{}, error) {
	// get disk type members.
	diskTypes := enumor.GetDiskTypeMembers()
	// convert to meta.CodeNameItem slice.
	rst := make([]meta.CodeNameItem, 0, len(diskTypes))
	for _, diskType := range diskTypes {
		rst = append(rst, meta.CodeNameItem{
			Code: diskType,
			Name: diskType.Name(),
		})
	}
	return rst, nil
}
