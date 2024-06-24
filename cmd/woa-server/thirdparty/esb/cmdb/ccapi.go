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

// Package cmdb provides client to interact with cc 3.0 api
package cmdb

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"hcm/cmd/woa-server/common"
	"hcm/cmd/woa-server/common/querybuilder"
	"hcm/pkg/cc"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Client cc api interface
type Client interface {
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
	// ListBizHost gets certain business host info in cc 3.0
	ListBizHost(ctx context.Context, header http.Header, req *ListBizHostReq) (*ListBizHostResp, error)
	// ListHost gets hosts info in cc 3.0
	ListHost(ctx context.Context, header http.Header, req *ListHostReq) (*ListHostResp, error)
	// FindHostBizRelation find host business relations, limit 500
	FindHostBizRelation(ctx context.Context, header http.Header, req *HostBizRelReq) (*HostBizRelResp, error)
	// SearchBiz search business, limit 200
	SearchBiz(ctx context.Context, header http.Header, req *SearchBizReq) (*SearchBizResp, error)
	// SearchBizByUser search authorized business by user
	SearchBizByUser(ctx context.Context, header http.Header, req *SearchBizReq, user string) (*SearchBizResp, error)
	// SearchModule search module
	SearchModule(ctx context.Context, header http.Header, req *SearchModuleReq) (*SearchModuleResp, error)
	// GetBizInternalModule get business's internal module
	GetBizInternalModule(ctx context.Context, header http.Header, req *GetBizInternalModuleReq) (*BizInternalModuleResp,
		error)
	GetBizRecycleModuleID(ctx context.Context, header http.Header, bizID int64) (int64, error)
	// Hosts2CrTransit transfer hosts to given business's CR transit module in CMDB
	Hosts2CrTransit(ctx context.Context, header http.Header, req *CrTransitReq) (*CrTransitResp, error)
	// HostsCrTransit2Idle transfer hosts to given business's idle module in CMDB
	HostsCrTransit2Idle(ctx context.Context, header http.Header, req *CrTransitIdleReq) (*CrTransitResp, error)
	// SearchBizBelonging search cmdb business belonging.
	SearchBizBelonging(ctx context.Context, header http.Header, params *SearchBizBelongingParams) (
		*SearchBizBelongingRst, error)
	// GetHostInfoByIP get host info by ip in CMDB
	GetHostInfoByIP(ctx context.Context, header http.Header, ip string, bkCloudID int) (*HostInfo, error)
	// GetHostInfoByHostID get host info by host id in CMDB
	GetHostInfoByHostID(ctx context.Context, header http.Header, bkHostID int64) (*HostInfo, error)
}

// ccCli cc api interface implementation
type ccCli struct {
	client rest.ClientInterface
	opts   *cc.Esb
}

// NewClient new cc client
func NewClient(client rest.ClientInterface, opts *cc.Esb) Client {
	return &ccCli{client: client, opts: opts}
}

func (c *ccCli) getAuthHeader() (string, string) {
	key := "X-Bkapi-Authorization"
	val := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}", c.opts.AppCode,
		c.opts.AppSecret, c.opts.User)

	return key, val
}

func (c *ccCli) getAuthHeaderWithUser(user string) (string, string) {
	key := "X-Bkapi-Authorization"
	val := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}", c.opts.AppCode,
		c.opts.AppSecret, user)

	return key, val
}

// AddHost adds host to cc 3.0, once 10 hosts at most
func (c *ccCli) AddHost(ctx context.Context, header http.Header, req *AddHostReq) (*AddHostResp, error) {
	subPath := "/api/c/compapi/v2/cc/add_host_from_cmpy"
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
func (c *ccCli) TransferHost(ctx context.Context, header http.Header, req *TransferHostReq) (*TransferHostResp, error) {
	subPath := "/api/c/compapi/v2/cc/transfer_host_to_another_biz"
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

	return resp, err
}

// GetHostId gets host id by ip in cc 3.0
func (c *ccCli) GetHostId(ctx context.Context, header http.Header, ip string) (int64, error) {
	subPath := "/api/c/compapi/v2/cc/list_hosts_without_biz"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	req := &ListHostReq{
		HostPropertyFilter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    common.BKHostInnerIPField,
						Operator: querybuilder.OperatorEqual,
						Value:    ip,
					},
				},
			},
		},
		Fields: []string{
			common.BKHostIDField,
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

	return resp.Data.Info[0].BkHostId, nil
}

// UpdateHosts update host info in cc 3.0
func (c *ccCli) UpdateHosts(ctx context.Context, header http.Header, req *UpdateHostsReq) (*UpdateHostsResp, error) {
	subPath := "/api/c/compapi/v2/cc/batch_update_host"
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
func (c *ccCli) GetHostBizIds(ctx context.Context, header http.Header, hostIds []int64) (map[int64]int64, error) {
	subPath := "/api/c/compapi/v2/cc/find_host_biz_relations"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	result := make(map[int64]int64)
	start := 0
	end := len(hostIds)
	if len(hostIds) > common.BKMaxInstanceLimit {
		end = common.BKMaxInstanceLimit
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

		if len(resp.Data) < common.BKMaxInstanceLimit {
			break
		}

		start = end
		if end+common.BKMaxInstanceLimit > len(hostIds) {
			end = len(hostIds)
			continue
		}

		end += common.BKMaxInstanceLimit
	}

	return result, nil
}

// ListBizHost gets certain business host info in cc 3.0, limit 500
func (c *ccCli) ListBizHost(ctx context.Context, header http.Header, req *ListBizHostReq) (*ListBizHostResp, error) {
	subPath := "/api/c/compapi/v2/cc/list_biz_hosts"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(ListBizHostResp)
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

// ListHost gets hosts info in cc 3.0, limit 500
func (c *ccCli) ListHost(ctx context.Context, header http.Header, req *ListHostReq) (*ListHostResp, error) {
	subPath := "/api/c/compapi/v2/cc/list_hosts_without_biz"
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

// FindHostBizRelation find host business relations, limit 500
func (c *ccCli) FindHostBizRelation(ctx context.Context, header http.Header, req *HostBizRelReq) (*HostBizRelResp,
	error) {

	subPath := "/api/c/compapi/v2/cc/find_host_biz_relations"
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
func (c *ccCli) SearchBiz(ctx context.Context, header http.Header, req *SearchBizReq) (*SearchBizResp, error) {
	subPath := "/api/c/compapi/v2/cc/search_business"
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
func (c *ccCli) SearchBizByUser(ctx context.Context, header http.Header, req *SearchBizReq, user string) (
	*SearchBizResp, error) {

	subPath := "/api/c/compapi/v2/cc/search_business"
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

// SearchModule search module
func (c *ccCli) SearchModule(ctx context.Context, header http.Header, req *SearchModuleReq) (*SearchModuleResp, error) {
	subPath := "/api/c/compapi/v2/cc/search_module"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(SearchModuleResp)
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
func (c *ccCli) GetBizInternalModule(ctx context.Context, header http.Header, req *GetBizInternalModuleReq) (
	*BizInternalModuleResp, error) {

	subPath := "/api/c/compapi/v2/cc/get_biz_internal_module"
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
		return nil, err
	}

	return resp, nil
}

// GetBizRecycleModuleID get business recycle module ID
func (c *ccCli) GetBizRecycleModuleID(ctx context.Context, header http.Header, bizID int64) (int64, error) {
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
		if module.Default == DftModuleRecycle {
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
func (c *ccCli) Hosts2CrTransit(ctx context.Context, header http.Header, req *CrTransitReq) (*CrTransitResp, error) {
	subPath := "/api/c/compapi/v2/cc/hosts_to_cr_transit"
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
func (c *ccCli) HostsCrTransit2Idle(ctx context.Context, header http.Header, req *CrTransitIdleReq) (*CrTransitResp,
	error) {

	subPath := "/api/c/compapi/v2/cc/hosts_cr_transit_to_idle"
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

// SearchBizBelonging search cmdb business belonging.
func (c *ccCli) SearchBizBelonging(ctx context.Context, header http.Header, req *SearchBizBelongingParams) (
	*SearchBizBelongingRst, error) {

	subPath := "/api/c/compapi/v2/cc/search_cost_info_relation"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	resp := new(SearchBizBelongingRst)
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
func (c *ccCli) GetHostInfoByIP(ctx context.Context, header http.Header, ip string, bkCloudID int) (*HostInfo, error) {
	subPath := "/api/c/compapi/v2/cc/list_hosts_without_biz"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	req := &ListHostReq{
		HostPropertyFilter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    common.BKHostInnerIPField,
						Operator: querybuilder.OperatorEqual,
						Value:    ip,
					},
					querybuilder.AtomRule{
						Field:    common.BKCloudIDField,
						Operator: querybuilder.OperatorEqual,
						Value:    bkCloudID,
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
		logs.Errorf("failed to get host info by ip, ip: %s, bkCloudID: %d, err: %v, subPath: %s",
			ip, bkCloudID, err, subPath)
		return nil, err
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
func (c *ccCli) GetHostInfoByHostID(ctx context.Context, header http.Header, bkHostID int64) (*HostInfo, error) {
	subPath := "/api/c/compapi/v2/cc/list_hosts_without_biz"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)

	req := &ListHostReq{
		HostPropertyFilter: &querybuilder.QueryFilter{
			Rule: querybuilder.CombinedRule{
				Condition: querybuilder.ConditionAnd,
				Rules: []querybuilder.Rule{
					querybuilder.AtomRule{
						Field:    common.BKHostIDField,
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
