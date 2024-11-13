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

	synctcloud "hcm/cmd/hc-service/logics/res-sync/tcloud"
	"hcm/cmd/hc-service/service/capability"
	adcore "hcm/pkg/adaptor/types/core"
	typecvm "hcm/pkg/adaptor/types/cvm"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	protocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
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
	h.Add("BatchStartTCloudZiyanCvm", http.MethodPost, "/vendors/tcloud-ziyan/cvms/batch/start",
		svc.BatchStartTCloudZiyanCvm)
	h.Add("BatchStopTCloudZiyanCvm", http.MethodPost, "/vendors/tcloud-ziyan/cvms/batch/stop",
		svc.BatchStopTCloudZiyanCvm)
	h.Add("BatchRebootTCloudZiyanCvm", http.MethodPost, "/vendors/tcloud-ziyan/cvms/batch/reboot",
		svc.BatchRebootTCloudZiyanCvm)
	h.Add("BatchResetTCloudZiyanCvm", http.MethodPost, "/vendors/tcloud-ziyan/cvms/reset", svc.BatchResetTCloudZiyanCvm)
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

	cvmWithCount, err := tziyan.ListCvmWithCount(cts.Kit, &typecvm.ListCvmWithCountOption{
		Region:   req.Region,
		CloudIDs: req.CvmIDs,
		SGIDs:    req.SGIDs,
		Page: &adcore.TCloudPage{
			Offset: uint64(req.Page.Start),
			Limit:  uint64(req.Page.Limit),
		},
	})
	if err != nil {
		logs.Errorf("fail to list cvm with count, err: %v, req: %+v ,rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	cmvList := slice.Map(cvmWithCount.Cvms, func(c typecvm.TCloudCvm) corecvm.Cvm[corecvm.TCloudCvmExtension] {
		return convTCloudCvm(c, req.AccountID, req.Region)
	})
	return types.ListResult[corecvm.Cvm[corecvm.TCloudCvmExtension]]{Count: uint64(cvmWithCount.TotalCount),
		Details: cmvList}, nil
}

// BatchStartTCloudZiyanCvm ...
func (svc *cvmSvc) BatchStartTCloudZiyanCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchStartReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudStartOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
	}
	if err = client.StartCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to start tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchStopTCloudZiyanCvm ...
func (svc *cvmSvc) BatchStopTCloudZiyanCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchStopReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudStopOption{
		Region:      req.Region,
		CloudIDs:    cloudIDs,
		StopType:    req.StopType,
		StoppedMode: req.StoppedMode,
	}
	if err = client.StopCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to stop tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// BatchRebootTCloudZiyanCvm ...
func (svc *cvmSvc) BatchRebootTCloudZiyanCvm(cts *rest.Contexts) (interface{}, error) {
	req := new(protocvm.TCloudBatchRebootReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Fields: []string{"cloud_id"},
		Filter: tools.ContainersExpression("id", req.IDs),
		Page:   core.NewDefaultBasePage(),
	}
	listResp, err := svc.dataCli.Global.Cvm.ListCvm(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("request dataservice list tcloud cvm failed, err: %v, ids: %v, rid: %s", err, req.IDs, cts.Kit.Rid)
		return nil, err
	}

	cloudIDs := make([]string, 0, len(listResp.Details))
	for _, one := range listResp.Details {
		cloudIDs = append(cloudIDs, one.CloudID)
	}

	client, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	opt := &typecvm.TCloudRebootOption{
		Region:   req.Region,
		CloudIDs: cloudIDs,
		StopType: req.StopType,
	}
	if err = client.RebootCvm(cts.Kit, opt); err != nil {
		logs.Errorf("request adaptor to reboot tcloud cvm failed, err: %v, opt: %v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	syncClient := synctcloud.NewClient(svc.dataCli, client)

	params := &synctcloud.SyncBaseParams{
		AccountID: req.AccountID,
		Region:    req.Region,
		CloudIDs:  cloudIDs,
	}

	_, err = syncClient.Cvm(cts.Kit, params, &synctcloud.SyncCvmOption{})
	if err != nil {
		logs.Errorf("sync tcloud cvm failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func convTCloudCvm(c typecvm.TCloudCvm, accountID, region string) corecvm.Cvm[corecvm.TCloudCvmExtension] {
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

// BatchResetTCloudZiyanCvm 重装系统
func (svc *cvmSvc) BatchResetTCloudZiyanCvm(cts *rest.Contexts) (any, error) {
	req := new(protocvm.TCloudBatchResetCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := svc.ad.TCloudZiyan(cts.Kit, req.AccountID)
	if err != nil {
		return nil, err
	}

	for _, cloudID := range req.CloudIDs {
		opt := &typecvm.ResetInstanceOption{
			Region:   req.Region,
			CloudID:  cloudID,
			ImageID:  req.ImageID,
			Password: req.Password,
		}
		if _, err = client.ResetCvmInstance(cts.Kit, opt); err != nil {
			logs.Errorf("request adaptor to tcloud ziyan reset cvm instance failed, err: %v, opt: %+v, cloudID: %s, "+
				"rid: %s", err, cvt.PtrToVal(req), cloudID, cts.Kit.Rid)
			return nil, err
		}
	}

	return nil, nil
}
