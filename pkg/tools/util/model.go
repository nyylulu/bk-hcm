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

// Package util ...
package util

import (
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
)

// AddModelBizIDCondition add model bizID condition according to bizID value
func AddModelBizIDCondition(cond mapstr.MapStr, modelBizID int64) {
	var modelBizIDOrCondArr []mapstr.MapStr
	if modelBizID > 0 {
		// special business model and global shared model
		modelBizIDOrCondArr = []mapstr.MapStr{
			{pkg.BKAppIDField: modelBizID},
			{pkg.BKAppIDField: 0},
			{pkg.BKAppIDField: mapstr.MapStr{pkg.BKDBExists: false}},
		}
	} else {
		// global shared model
		modelBizIDOrCondArr = []mapstr.MapStr{
			{pkg.BKAppIDField: 0},
			{pkg.BKAppIDField: mapstr.MapStr{pkg.BKDBExists: false}},
		}
	}

	if _, exists := cond[pkg.BKDBOR]; !exists {
		cond[pkg.BKDBOR] = modelBizIDOrCondArr
	} else {
		andCondArr := []map[string]interface{}{
			{pkg.BKDBOR: modelBizIDOrCondArr},
		}

		andCond, exists := cond[pkg.BKDBAND]
		if !exists {
			cond[pkg.BKDBAND] = andCondArr
		} else {
			cond[pkg.BKDBAND] = append(andCondArr, map[string]interface{}{pkg.BKDBAND: andCond})
		}
	}
	delete(cond, pkg.BKAppIDField)
}
