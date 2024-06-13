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

package cvmapi

import (
	"context"
	"hcm/pkg/cc"
	"net/http"

	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// CVMClientInterface cvm api interface
type CVMClientInterface interface {
	// CreateCvmOrder creates cvm order
	CreateCvmOrder(ctx context.Context, header http.Header, req *OrderCreateReq) (*OrderCreateResp, error)
	// QueryCvmOrders query cvm orders
	QueryCvmOrders(ctx context.Context, header http.Header, req *OrderQueryReq) (*OrderQueryResp, error)
	// QueryCvmInstances query cvm instances
	QueryCvmInstances(ctx context.Context, header http.Header, req *InstanceQueryReq) (*InstanceQueryResp, error)
	// QueryCvmCapacity query cvm capacity
	QueryCvmCapacity(ctx context.Context, header http.Header, req *CapacityReq) (*CapacityResp, error)
	// QueryCvmVpc query cvm subnet info
	QueryCvmVpc(ctx context.Context, header http.Header, req *VpcReq) (*VpcResp, error)
	// QueryCvmSubnet query cvm subnet info
	QueryCvmSubnet(ctx context.Context, header http.Header, req *SubnetReq) (*SubnetResp, error)
	// QueryCvmCbsPlans query cvm and cbs plan info
	QueryCvmCbsPlans(ctx context.Context, header http.Header, req *CvmCbsPlanQueryReq) (*CvmCbsPlanQueryResp, error)
	// AdjustCvmCbsPlans adjust cvm and cbs plan info
	AdjustCvmCbsPlans(ctx context.Context, header http.Header, req *CvmCbsPlanAdjustReq) (*CvmCbsPlanAdjustResp, error)
	// AddCvmCbsPlan add cvm and cbs plan order
	AddCvmCbsPlan(ctx context.Context, header http.Header, req *AddCvmCbsPlanReq) (*AddCvmCbsPlanResp, error)
	// QueryPlanOrder query cvm and cbs plan order
	QueryPlanOrder(ctx context.Context, header http.Header, req *QueryPlanOrderReq) (*QueryPlanOrderResp, error)
	// CreateCvmReturnOrder creates cvm return order
	CreateCvmReturnOrder(ctx context.Context, header http.Header, req *ReturnReq) (*OrderCreateResp, error)
	// QueryCvmReturnOrders query cvm return order status
	QueryCvmReturnOrders(ctx context.Context, header http.Header, req *OrderQueryReq) (*ReturnQueryResp, error)
	// QueryCvmReturnDetail query cvm return order detail
	QueryCvmReturnDetail(ctx context.Context, header http.Header, req *ReturnDetailReq) (*ReturnDetailResp, error)
	// GetCvmProcess check if cvm is in any process like "退回"
	GetCvmProcess(ctx context.Context, header http.Header, req *GetCvmProcessReq) (*GetCvmProcessResp, error)
	// GetErpProcess check if physical machine is in any process like "退回"
	GetErpProcess(ctx context.Context, header http.Header, req *GetErpProcessReq) (*GetErpProcessResp, error)
}

// NewCVMClientInterface creates a cvm api instance
func NewCVMClientInterface(opts cc.CVMCli, reg prometheus.Registerer) (CVMClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "cvm api",
			servers: []string{opts.CvmApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	cvm := &cvmApi{
		client: rest.NewClient(c, "/"),
	}

	return cvm, nil
}

// cvmApi cvm api interface implementation
type cvmApi struct {
	client rest.ClientInterface
}

// CreateCvmOrder creates cvm order
func (c *cvmApi) CreateCvmOrder(ctx context.Context, header http.Header, req *OrderCreateReq) (*OrderCreateResp,
	error) {

	subPath := "/apply/api/cvm"
	resp := new(OrderCreateResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmOrders query cvm orders
func (c *cvmApi) QueryCvmOrders(ctx context.Context, header http.Header, req *OrderQueryReq) (*OrderQueryResp, error) {
	subPath := "/apply/api/cvm"
	resp := new(OrderQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmInstances query cvm instances
func (c *cvmApi) QueryCvmInstances(ctx context.Context, header http.Header, req *InstanceQueryReq) (*InstanceQueryResp,
	error) {

	subPath := "/apply/api/cvm"
	resp := new(InstanceQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmCapacity query cvm inventory
func (c *cvmApi) QueryCvmCapacity(ctx context.Context, header http.Header, req *CapacityReq) (*CapacityResp, error) {
	subPath := "/capacity/api/queryApplyCapacity"
	resp := new(CapacityResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmVpc query cvm subnet info
func (c *cvmApi) QueryCvmVpc(ctx context.Context, header http.Header, req *VpcReq) (*VpcResp, error) {
	subPath := "/apply/api/cvm"
	resp := new(VpcResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmSubnet query cvm subnet info
func (c *cvmApi) QueryCvmSubnet(ctx context.Context, header http.Header, req *SubnetReq) (*SubnetResp, error) {
	subPath := "/apply/api/cvm"
	resp := new(SubnetResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmCbsPlans query cvm and cbs plans
func (c *cvmApi) QueryCvmCbsPlans(ctx context.Context, header http.Header, req *CvmCbsPlanQueryReq) (
	*CvmCbsPlanQueryResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(CvmCbsPlanQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// AdjustCvmCbsPlans adjust cvm and cbs plans
func (c *cvmApi) AdjustCvmCbsPlans(ctx context.Context, header http.Header, req *CvmCbsPlanAdjustReq) (
	*CvmCbsPlanAdjustResp, error) {

	subPath := "/yunti-demand/external"
	resp := new(CvmCbsPlanAdjustResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// AddCvmCbsPlan add cvm and cbs plan order
func (c *cvmApi) AddCvmCbsPlan(ctx context.Context, header http.Header, req *AddCvmCbsPlanReq) (*AddCvmCbsPlanResp,
	error) {

	subPath := "/yunti-demand/external"
	resp := new(AddCvmCbsPlanResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryPlanOrder query cvm and cbs plan order
func (c *cvmApi) QueryPlanOrder(ctx context.Context, header http.Header, req *QueryPlanOrderReq) (*QueryPlanOrderResp,
	error) {

	subPath := "/yunti-demand/external"
	resp := new(QueryPlanOrderResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// CreateCvmReturnOrder creates cvm return order
func (c *cvmApi) CreateCvmReturnOrder(ctx context.Context, header http.Header, req *ReturnReq) (*OrderCreateResp,
	error) {

	subPath := "/apply/api"
	resp := new(OrderCreateResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmReturnOrders query cvm return order status
func (c *cvmApi) QueryCvmReturnOrders(ctx context.Context, header http.Header, req *OrderQueryReq) (*ReturnQueryResp,
	error) {

	subPath := "/apply/api"
	resp := new(ReturnQueryResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// QueryCvmReturnDetail query cvm return order detail
func (c *cvmApi) QueryCvmReturnDetail(ctx context.Context, header http.Header, req *ReturnDetailReq) (*ReturnDetailResp,
	error) {

	subPath := "/apply/api"
	resp := new(ReturnDetailResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetCvmProcess check if cvm is in any process like "退回"
func (c *cvmApi) GetCvmProcess(ctx context.Context, header http.Header, req *GetCvmProcessReq) (*GetCvmProcessResp,
	error) {

	subPath := "/operation/api/"
	resp := new(GetCvmProcessResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}

// GetErpProcess check if physical machine is in any process like "退回"
func (c *cvmApi) GetErpProcess(ctx context.Context, header http.Header, req *GetErpProcessReq) (*GetErpProcessResp,
	error) {

	subPath := "/operation/api/"
	resp := new(GetErpProcessResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithParam(CvmApiKey, CvmApiKeyVal).
		WithHeaders(header).
		Do().
		Into(resp)

	return resp, err
}
