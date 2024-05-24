/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package cvm

import (
	"net/http"

	"hcm/cmd/hc-service/service/capability"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/adaptor/types/cvm"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	tcvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

func (svc *cvmSvc) initTCloudZiyanCvmService(cap *capability.Capability) {
	h := rest.NewHandler()

	h.Add("QueryTCloudZiyanCvm", http.MethodPost,
		"/vendors/tcloud-ziyan/cvms/query", svc.BatchQueryTCloudZiyanCVM)
	h.Load(cap.WebService)
}

// BatchQueryTCloudZiyanCVM 到云上查询cvm
func (svc *cvmSvc) BatchQueryTCloudZiyanCVM(cts *rest.Contexts) (any, error) {
	req := new(corecvm.QueryCloudCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	tziyan, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	cvmWithCount, err := tziyan.ListCvmWithCount(cts.Kit, &cvm.ListCvmWithCountOption{
		Region:   req.Region,
		CloudIDs: req.CvmIDs,
		SGIDs:    req.SGIDs,
		Page: &core.TCloudPage{
			Offset: uint64(req.Page.Start),
			Limit:  uint64(req.Page.Limit),
		},
	})
	if err != nil {
		logs.Errorf("fail to list cvm with count, err: %v, req: %+v ,rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	cmvList := slice.Map(cvmWithCount.Cvms, func(c cvm.TCloudCvm) corecvm.Cvm[corecvm.TCloudCvmExtension] {
		return convTCloudCvm(c, req.AccountID, req.Region)
	})
	return types.ListResult[corecvm.Cvm[corecvm.TCloudCvmExtension]]{Count: uint64(cvmWithCount.TotalCount),
		Details: cmvList}, nil
}

func convTCloudCvm(c cvm.TCloudCvm, accountID, region string) corecvm.Cvm[corecvm.TCloudCvmExtension] {
	var cloudVpcIDs, cloudSubnetIDs []string
	if c.VirtualPrivateCloud != nil {
		cloudVpcIDs = append(cloudVpcIDs, cvt.PtrToVal(c.VirtualPrivateCloud.VpcId))
		cloudSubnetIDs = append(cloudSubnetIDs, cvt.PtrToVal(c.VirtualPrivateCloud.SubnetId))
	}
	baseCvm := corecvm.BaseCvm{
		CloudID:              c.GetCloudID(),
		Name:                 cvt.PtrToVal(c.InstanceName),
		Vendor:               enumor.TCloudZiyan,
		AccountID:            accountID,
		Region:               region,
		Zone:                 cvt.PtrToVal(c.Placement.Zone),
		CloudVpcIDs:          cloudVpcIDs,
		CloudSubnetIDs:       cloudSubnetIDs,
		OsName:               cvt.PtrToVal(c.OsName),
		Status:               cvt.PtrToVal(c.InstanceState),
		PrivateIPv4Addresses: cvt.PtrToSlice(c.PrivateIpAddresses),
		PublicIPv4Addresses:  cvt.PtrToSlice(c.PublicIpAddresses),
		PrivateIPv6Addresses: cvt.PtrToSlice(c.IPv6Addresses),
		MachineType:          cvt.PtrToVal(c.InstanceType),
		CloudImageID:         cvt.PtrToVal(c.ImageId),
		CloudCreatedTime:     cvt.PtrToVal(c.CreatedTime),
	}
	var internetAccessible *corecvm.TCloudInternetAccessible
	if c.InternetAccessible != nil {
		internetAccessible = &corecvm.TCloudInternetAccessible{
			InternetChargeType:      c.InternetAccessible.InternetChargeType,
			InternetMaxBandwidthOut: c.InternetAccessible.InternetMaxBandwidthOut,
			PublicIPAssigned:        c.InternetAccessible.PublicIpAssigned,
			CloudBandwidthPackageID: c.InternetAccessible.BandwidthPackageId,
		}
	}

	return corecvm.Cvm[corecvm.TCloudCvmExtension]{
		BaseCvm: baseCvm,
		Extension: &corecvm.TCloudCvmExtension{
			CloudSecurityGroupIDs: cvt.PtrToSlice(c.SecurityGroupIds),
			Placement:             &corecvm.TCloudPlacement{CloudProjectID: cvt.PtrToVal(c.Placement).ProjectId},
			InstanceChargeType:    c.InstanceChargeType,
			Cpu:                   c.CPU,
			Memory:                c.Memory,
			CloudSystemDiskID:     cvt.PtrToVal(c.SystemDisk).DiskId,
			CloudDataDiskIDs: slice.Map(c.DataDisks,
				func(dd *tcvm.DataDisk) string { return cvt.PtrToVal(dd.DiskId) }),
			InternetAccessible: internetAccessible,

			VirtualPrivateCloud: &corecvm.TCloudVirtualPrivateCloud{
				AsVpcGateway: c.VirtualPrivateCloud.AsVpcGateway,
			},
			RenewFlag:             c.RenewFlag,
			StopChargingMode:      c.StopChargingMode,
			UUID:                  c.Uuid,
			IsolatedSource:        c.IsolatedSource,
			DisableApiTermination: c.DisableApiTermination,
		},
	}
}
