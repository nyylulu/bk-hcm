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

package config

import (
	"hcm/cmd/woa-server/model/config"
	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// CvmImageIf provides management interface for operations of cvm image config
type CvmImageIf interface {
	// GetCvmImage get cvm image type config list
	GetCvmImage(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetCvmImageResult, error)
	// CreateCvmImage creates cvm image type config
	CreateCvmImage(kt *kit.Kit, input *types.CvmImage) (mapstr.MapStr, error)
	// UpdateCvmImage updates cvm image type config
	UpdateCvmImage(kt *kit.Kit, instId int64, input *mapstr.MapStr) error
	// DeleteCvmImage deletes cvm image type config
	DeleteCvmImage(kt *kit.Kit, instId int64) error
}

// NewCvmImageOp creates a cvm image interface
func NewCvmImageOp() CvmImageIf {
	return &cvmImage{}
}

type cvmImage struct {
}

// GetCvmImage get cvm image type config list
func (i *cvmImage) GetCvmImage(kt *kit.Kit, cond *mapstr.MapStr) (*types.GetCvmImageResult, error) {
	insts, err := config.Operation().CvmImage().FindManyCvmImage(kt.Ctx, cond)
	if err != nil {
		return nil, err
	}

	// remove duplicate image
	imageMap := make(map[string]*types.CvmImage)
	imageList := make([]*types.CvmImage, 0)
	for _, inst := range insts {
		if _, ok := imageMap[inst.ImageId]; !ok {
			image := &types.CvmImage{
				ImageId:   inst.ImageId,
				ImageName: inst.ImageName,
			}
			imageMap[inst.ImageId] = image
			imageList = append(imageList, image)
		}
	}

	rst := &types.GetCvmImageResult{
		Count: int64(len(imageList)),
		Info:  imageList,
	}

	return rst, nil
}

// CreateCvmImage creates cvm image type config
func (i *cvmImage) CreateCvmImage(kt *kit.Kit, input *types.CvmImage) (mapstr.MapStr, error) {
	id, err := config.Operation().CvmImage().NextSequence(kt.Ctx)
	if err != nil {
		logs.Errorf("failed to create cvm image, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	instId := int64(id)

	input.BkInstId = instId
	if err := config.Operation().CvmImage().CreateCvmImage(kt.Ctx, input); err != nil {
		logs.Errorf("failed to create cvm image, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rst := mapstr.MapStr{
		"id": instId,
	}

	return rst, nil
}

// UpdateCvmImage updates cvm image type config
func (i *cvmImage) UpdateCvmImage(kt *kit.Kit, instId int64, input *mapstr.MapStr) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().CvmImage().UpdateCvmImage(kt.Ctx, filter, input); err != nil {
		logs.Errorf("failed to update cvm image, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// DeleteCvmImage deletes cvm image type config
func (i *cvmImage) DeleteCvmImage(kt *kit.Kit, instId int64) error {
	filter := &mapstr.MapStr{
		"id": instId,
	}

	if err := config.Operation().CvmImage().DeleteCvmImage(kt.Ctx, filter); err != nil {
		logs.Errorf("failed to delete cvm image, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}
