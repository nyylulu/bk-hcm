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

package cmdb

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"hcm/pkg"
	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/esb/types"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/querybuilder"
	"hcm/pkg/tools/utils"
)

// Client is an esb client to request cmdb.
type Client interface {
	SearchBusiness(kt *kit.Kit, params *SearchBizParams) (*SearchBizResult, error)
	SearchCloudArea(kt *kit.Kit, params *SearchCloudAreaParams) (*SearchCloudAreaResult, error)
	AddCloudHostToBiz(kt *kit.Kit, params *AddCloudHostToBizParams) (*BatchCreateResult, error)
	DeleteCloudHostFromBiz(kt *kit.Kit, params *DeleteCloudHostFromBizParams) error
	ListBizHost(kt *kit.Kit, params *ListBizHostParams) (*ListBizHostResult, error)
	GetBizBriefCacheTopo(kt *kit.Kit, params *GetBizBriefCacheTopoParams) (*GetBizBriefCacheTopoResult, error)
	FindHostTopoRelation(kt *kit.Kit, params *FindHostTopoRelationParams) (*HostTopoRelationResult, error)
	SearchModule(kt *kit.Kit, params *SearchModuleParams) (*ModuleInfoResult, error)
	// SearchBizBelonging search cmdb business belonging.
	SearchBizBelonging(kt *kit.Kit, params *SearchBizBelongingParams) (*[]SearchBizBelonging, error)
	ResourceWatch(kt *kit.Kit, params *WatchEventParams) (*WatchEventResult, error)
	FindHostBizRelations(kt *kit.Kit, params *HostModuleRelationParams) (*[]HostTopoRelation, error)
	ListHost(ctx context.Context, header http.Header, req *ListHostReq) (*ListHostResp, error)

	// AddHost adds host to cc 3.0, once 10 hosts at most
	AddHost(ctx context.Context, header http.Header, req *AddHostReq) (*AddHostResp, error)
	// TransferHost transfer host to another business in cc 3.0
	TransferHost(ctx context.Context, header http.Header, req *TransferHostReq) (*TransferHostResp, error)
	// GetHostId gets host id by ip in cc 3.0
	GetHostId(ctx context.Context, header http.Header, ip string) (int64, error)
	// UpdateHosts update host info in cc 3.0
	UpdateHosts(ctx context.Context, header http.Header, req *UpdateHostsReq) (*UpdateHostsResp, error)
	// GetHostBizIds gets host biz id by host id in cc 3.0
	GetHostBizIds(ctx context.Context, header http.Header, hostId []int64) (map[int64]int64, error)
	// FindHostBizRelation find host business relations, limit 500
	FindHostBizRelation(ctx context.Context, header http.Header, req *HostBizRelReq) (*HostBizRelResp, error)
	// SearchBiz search business, limit 200
	SearchBiz(ctx context.Context, header http.Header, req *SearchBizReq) (*SearchBizResp, error)
	// SearchBizByUser search authorized business by user
	SearchBizByUser(ctx context.Context, header http.Header, req *SearchBizReq, user string) (*SearchBizResp, error)
	// GetBizInternalModule get business's internal module
	GetBizInternalModule(ctx context.Context, header http.Header, req *GetBizInternalModuleReq) (*BizInternalModuleResp,
		error)
	GetBizRecycleModuleID(ctx context.Context, header http.Header, bizID int64) (int64, error)
	// Hosts2CrTransit transfer hosts to given business's CR transit module in CMDB
	Hosts2CrTransit(ctx context.Context, header http.Header, req *CrTransitReq) (*CrTransitResp, error)
	// HostsCrTransit2Idle transfer hosts to given business's idle module in CMDB
	HostsCrTransit2Idle(ctx context.Context, header http.Header, req *CrTransitIdleReq) (*CrTransitResp, error)
	// GetHostInfoByIP get host info by ip in CMDB
	GetHostInfoByIP(ctx context.Context, header http.Header, ip string, bkCloudID int) (*Host, error)
	// GetHostInfoByHostID get host info by host id in CMDB
	GetHostInfoByHostID(ctx context.Context, header http.Header, bkHostID int64) (*Host, error)
}

// NewClient initialize a new cmdb client
func NewClient(client rest.ClientInterface, config *cc.Esb) Client {
	return &cmdb{
		client: client,
		config: config,
	}
}

var _ Client = new(cmdb)

// cmdb is an esb client to request cmdb.
type cmdb struct {
	config *cc.Esb
	// http client instance
	client rest.ClientInterface
}

// SearchBusiness search business
func (c *cmdb) SearchBusiness(kt *kit.Kit, params *SearchBizParams) (*SearchBizResult, error) {

	return types.EsbCall[SearchBizParams, SearchBizResult](c.client, c.config, rest.POST, kt, params,
		"/cc/search_business/")
}

// SearchCloudArea search cmdb cloud area
func (c *cmdb) SearchCloudArea(kt *kit.Kit, params *SearchCloudAreaParams) (*SearchCloudAreaResult, error) {

	return types.EsbCall[SearchCloudAreaParams, SearchCloudAreaResult](c.client, c.config, rest.POST, kt, params,
		"/cc/search_cloud_area/")
}

// AddCloudHostToBiz add cmdb cloud host to biz.
func (c *cmdb) AddCloudHostToBiz(kt *kit.Kit, params *AddCloudHostToBizParams) (*BatchCreateResult, error) {

	return types.EsbCall[AddCloudHostToBizParams, BatchCreateResult](c.client, c.config, rest.POST, kt, params,
		"/cc/add_cloud_host_to_biz/")
}

// DeleteCloudHostFromBiz delete cmdb cloud host from biz.
func (c *cmdb) DeleteCloudHostFromBiz(kt *kit.Kit, params *DeleteCloudHostFromBizParams) error {
	_, err := types.EsbCall[DeleteCloudHostFromBizParams, struct{}](c.client, c.config, rest.POST, kt, params,
		"/cc/delete_cloud_host_from_biz/")
	return err
}

// ListBizHost list cmdb host in biz.
func (c *cmdb) ListBizHost(kt *kit.Kit, params *ListBizHostParams) (*ListBizHostResult, error) {

	return types.EsbCall[ListBizHostParams, ListBizHostResult](c.client, c.config, rest.POST, kt, params,
		"/cc/list_biz_hosts/")
}

// FindHostTopoRelation 获取主机拓扑
func (c *cmdb) FindHostTopoRelation(kt *kit.Kit, params *FindHostTopoRelationParams) (
	*HostTopoRelationResult, error) {

	return types.EsbCall[FindHostTopoRelationParams, HostTopoRelationResult](c.client, c.config, rest.POST, kt, params,
		"/cc/find_host_topo_relation/")
}

// SearchModule 查询模块信息
func (c *cmdb) SearchModule(kt *kit.Kit, params *SearchModuleParams) (*ModuleInfoResult, error) {

	return types.EsbCall[SearchModuleParams, ModuleInfoResult](c.client, c.config, rest.POST, kt, params,
		"/cc/search_module/")
}

// SearchBizBelonging search cmdb business belonging.
func (c *cmdb) SearchBizBelonging(kt *kit.Kit, params *SearchBizBelongingParams) (*[]SearchBizBelonging, error) {

	return types.EsbCall[SearchBizBelongingParams, []SearchBizBelonging](c.client, c.config, rest.POST, kt, params,
		"/cc/search_cost_info_relation/")
}

// ResourceWatch watch cmdb resource event.
func (c *cmdb) ResourceWatch(kt *kit.Kit, params *WatchEventParams) (*WatchEventResult, error) {
	return types.EsbCall[WatchEventParams, WatchEventResult](c.client, c.config, rest.POST, kt, params,
		"/cc/resource_watch/")
}

// FindHostBizRelations find host biz relations.
func (c *cmdb) FindHostBizRelations(kt *kit.Kit, params *HostModuleRelationParams) (*[]HostTopoRelation, error) {
	return types.EsbCall[HostModuleRelationParams, []HostTopoRelation](c.client, c.config, rest.POST, kt, params,
		"/cc/find_host_biz_relations/")
}

// ListHost gets hosts info in cc 3.0, limit 500
func (c *cmdb) ListHost(ctx context.Context, header http.Header, req *ListHostReq) (*ListHostResp, error) {
	subPath := "/cc/list_hosts_without_biz/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(ListHostResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// cc api functions

// AddHost adds host to cc 3.0, once 10 hosts at most
func (c *cmdb) AddHost(ctx context.Context, header http.Header, req *AddHostReq) (*AddHostResp, error) {
	subPath := "/cc/add_host_from_cmpy/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(AddHostResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// TransferHost transfer host to another business in cc 3.0
func (c *cmdb) TransferHost(ctx context.Context, header http.Header, req *TransferHostReq) (*TransferHostResp, error) {
	subPath := "/cc/transfer_host_to_another_biz/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(TransferHostResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("scheduler:cmdb:TransferHost:failed, err: %v, req: %+v", err, cvt.PtrToVal(req))
		return nil, err
	}

	return resp, err
}

// GetHostId gets host id by ip in cc 3.0
func (c *cmdb) GetHostId(ctx context.Context, header http.Header, ip string) (int64, error) {
	subPath := "/cc/list_hosts_without_biz/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	req := &ListHostReq{
		HostPropertyFilter: &QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    pkg.BKHostInnerIPField,
						Operator: querybuilder.OperatorEqual,
						Value:    ip,
					},
				},
			},
		},
		Fields: []string{
			pkg.BKHostIDField,
		},
		Page: BasePage{
			Start: 0,
			Limit: 1,
		},
	}

	resp := new(ListHostResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return -1, err
	}

	if !resp.Result || resp.Code != 0 {
		return -1, fmt.Errorf("failed to get host id, err: %s", resp.ErrMsg)
	}

	if len(resp.Data.Info) != 1 {
		return -1, fmt.Errorf("failed to get host id, for return data size %d not equal 1",
			len(resp.Data.Info))
	}

	return resp.Data.Info[0].BkHostID, nil
}

// UpdateHosts update host info in cc 3.0
func (c *cmdb) UpdateHosts(ctx context.Context, header http.Header, req *UpdateHostsReq) (*UpdateHostsResp, error) {
	subPath := "/cc/batch_update_host/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(UpdateHostsResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetHostBizIds gets host biz id by host id in cc 3.0
func (c *cmdb) GetHostBizIds(ctx context.Context, header http.Header, hostIds []int64) (map[int64]int64, error) {
	subPath := "/cc/find_host_biz_relations/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	result := make(map[int64]int64)
	start := 0
	end := len(hostIds)
	if len(hostIds) > pkg.BKMaxInstanceLimit {
		end = pkg.BKMaxInstanceLimit
	}

	for {
		req := &HostModuleRelationParameter{
			HostID: hostIds[start:end],
		}
		resp := new(HostModuleResp)
		err := c.client.Post().
			WithContext(ctx).
			Body(req).
			SubResourcef(subPath).
			WithHeaders(header).
			Do().
			Into(resp)

		if err != nil {
			return nil, err
		}

		if !resp.Result || resp.Code != 0 {
			return nil, fmt.Errorf("failed to get host biz id, err: %s", resp.ErrMsg)
		}

		if len(resp.Data) == 0 {
			return nil, fmt.Errorf("failed to get host biz id, for return data size 0")
		}

		for _, data := range resp.Data {
			result[data.HostID] = data.AppID
		}

		if end == len(hostIds) {
			break
		}

		start = end
		if end+pkg.BKMaxInstanceLimit > len(hostIds) {
			end = len(hostIds)
			continue
		}

		end += pkg.BKMaxInstanceLimit
	}

	return result, nil
}

// FindHostBizRelation find host business relations, limit 500
func (c *cmdb) FindHostBizRelation(ctx context.Context, header http.Header, req *HostBizRelReq) (*HostBizRelResp,
	error) {

	subPath := "/cc/find_host_biz_relations/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(HostBizRelResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SearchBiz search business, limit 200
func (c *cmdb) SearchBiz(ctx context.Context, header http.Header, req *SearchBizReq) (*SearchBizResp, error) {
	subPath := "/cc/search_business/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(SearchBizResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SearchBizByUser search authorized business by user
func (c *cmdb) SearchBizByUser(ctx context.Context, header http.Header, req *SearchBizReq, user string) (
	*SearchBizResp, error) {

	subPath := "/cc/search_business/"
	key, val := c.getAuthHeaderWithUser(user)
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(SearchBizResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetBizInternalModule get business's internal module
func (c *cmdb) GetBizInternalModule(ctx context.Context, header http.Header, req *GetBizInternalModuleReq) (
	*BizInternalModuleResp, error) {

	subPath := "/cc/get_biz_internal_module/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(BizInternalModuleResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("recycler:cvm:get:biz:intetnal:module:failed, req: %+v, err: %v", cvt.PtrToVal(req), err)
		return nil, err
	}

	return resp, nil
}

// GetBizRecycleModuleID get business recycle module ID
func (c *cmdb) GetBizRecycleModuleID(ctx context.Context, header http.Header, bizID int64) (int64, error) {
	req := &GetBizInternalModuleReq{
		BkBizID: bizID,
	}
	resp, err := c.GetBizInternalModule(ctx, header, req)
	if err != nil {
		return 0, fmt.Errorf("failed to get biz internal module, err: %v", err)
	}

	if resp.Result == false || resp.Code != 0 {
		return 0, fmt.Errorf("failed to get biz internal module, code: %d, msg: %s", resp.Code, resp.ErrMsg)
	}

	moduleID := int64(0)
	for _, module := range resp.Data.Module {
		if int(module.Default) == DftModuleRecycle {
			moduleID = module.BkModuleId
			break
		}
	}

	if moduleID <= 0 {
		return 0, errors.New("get no biz recycle module ID")
	}

	return moduleID, nil
}

// Hosts2CrTransit transfer hosts to given business's CR transit module in CMDB
func (c *cmdb) Hosts2CrTransit(ctx context.Context, header http.Header, req *CrTransitReq) (*CrTransitResp, error) {
	subPath := "/cc/hosts_to_cr_transit/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(CrTransitResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// HostsCrTransit2Idle transfer hosts to given business's idle module in CMDB
func (c *cmdb) HostsCrTransit2Idle(ctx context.Context, header http.Header, req *CrTransitIdleReq) (*CrTransitResp,
	error) {

	subPath := "/cc/hosts_cr_transit_to_idle/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(CrTransitResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetHostInfoByIP get host info by ip in cc 3.0(bkCloudID是管控区ID)
func (c *cmdb) GetHostInfoByIP(ctx context.Context, header http.Header, ip string, bkCloudID int) (*Host, error) {
	subPath := "/cc/list_hosts_without_biz/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	req := &ListHostReq{
		HostPropertyFilter: &QueryFilter{
			Rule: CombinedRule{
				Condition: ConditionAnd,
				Rules: []Rule{
					AtomRule{
						Field:    pkg.BKHostInnerIPField,
						Operator: OperatorEqual,
						Value:    ip,
					},
					AtomRule{
						Field:    pkg.BKCloudIDField,
						Operator: OperatorEqual,
						Value:    bkCloudID,
					},
				},
			},
		},
		Page: BasePage{Start: 0, Limit: 1},
	}

	// 新增重试机制
	resp := new(ListHostResp)
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, err
		}
		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// construct order status request
		err := c.client.Post().
			WithContext(ctx).
			Body(req).
			SubResourcef(subPath).
			WithHeaders(header).
			Do().
			Into(resp)
		return resp, err
	}

	// TODO: get retry strategy from config
	obj, err := utils.Retry(doFunc, checkFunc, 120, 5)
	if err != nil {
		logs.Errorf("failed to get host info by ip, ip: %s, bkCloudID: %d, err: %v", ip, bkCloudID, err)
		return nil, err
	}

	resp, ok := obj.(*ListHostResp)
	if !ok {
		return nil, fmt.Errorf("failed to get host info, resp is not ListHostResp, ip: %s, bkCloudID: %d, err: %v", ip,
			bkCloudID, err)
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("failed to get host info, ip: %s, bkCloudID: %d, errCode: %d, errMsg: %s",
			ip, bkCloudID, resp.Code, resp.ErrMsg)
	}

	if len(resp.Data.Info) != 1 {
		return nil, fmt.Errorf("failed to get host info by ip, ip: %s, bkCloudID: %d, for return "+
			"data size %d not equal 1", ip, bkCloudID, len(resp.Data.Info))
	}

	return resp.Data.Info[0], nil
}

// GetHostInfoByHostID get hosts info by host id in cc 3.0
func (c *cmdb) GetHostInfoByHostID(ctx context.Context, header http.Header, bkHostID int64) (*Host, error) {
	subPath := "/cc/list_hosts_without_biz/"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	req := &ListHostReq{
		HostPropertyFilter: &QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    pkg.BKHostIDField,
						Operator: querybuilder.OperatorEqual,
						Value:    bkHostID,
					},
				},
			},
		},
		Page: BasePage{Start: 0, Limit: 1},
	}

	resp := new(ListHostResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("failed to get host info by host id, bkHostID: %d, err: %v, subPath: %s", bkHostID, err, subPath)
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("failed to get host info by host id, bkHostID: %d, errCode: %d, errMsg: %s, subPath: %s",
			bkHostID, resp.Code, resp.ErrMsg, subPath)
	}

	if len(resp.Data.Info) != 1 {
		return nil, fmt.Errorf("failed to get host info by host id, bkHostID: %d, for return data size %d not equal 1",
			bkHostID, len(resp.Data.Info))
	}

	return resp.Data.Info[0], nil
}

func (c *cmdb) getAuthHeaderWithUser(user string) (string, string) {
	key := "X-Bkapi-Authorization"
	val := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}", c.config.AppCode,
		c.config.AppSecret, user)

	return key, val
}
