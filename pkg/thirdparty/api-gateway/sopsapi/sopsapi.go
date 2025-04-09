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

package sopsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"

	"github.com/prometheus/client_golang/prometheus"
)

// SopsClientInterface sops api interface
type SopsClientInterface interface {
	// CreateTask create sops task
	CreateTask(ctx context.Context, header http.Header, templateID, bkBizID int64, req *CreateTaskReq) (
		*CreateTaskResp, error)
	// StartTask start sops task
	StartTask(ctx context.Context, header http.Header, taskID, bkBizID int64) (*StartTaskResp, error)
	// GetTaskStatus get sops task status
	GetTaskStatus(ctx context.Context, header http.Header, taskID, bkBizID int64) (*GetTaskStatusResp, error)
	// GetTaskNodeDetail get sops task node detail
	GetTaskNodeDetail(ctx context.Context, header http.Header, taskID, bkBizID int64, nodeID string) (
		*GetTaskNodeDetailResp, error)
	// GetTaskNodeData get sops task node data
	GetTaskNodeData(kt *kit.Kit, header http.Header, taskID, bkBizID int64, nodeID string, subStacks string) (
		*GetTaskNodeDataRst, error)
	// GetTaskList get sops task list
	GetTaskList(ctx context.Context, header http.Header, bkBizID int64, keyword string) ([]*GetTaskListRst, error)
	// GetTaskDetail get sops task detail
	GetTaskDetail(kt *kit.Kit, header http.Header, taskID, bkBizID int64) (*GetTaskDetailDataResp, error)
}

// NewSopsClientInterface creates a sops api instance
func NewSopsClientInterface(opts cc.SopsCli, reg prometheus.Registerer) (SopsClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "sops api",
			servers: []string{opts.SopsApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	sopsCli := &sopsApi{
		client: rest.NewClient(c, "/"),
		opts:   opts,
	}

	return sopsCli, nil
}

// GetTaskList get sops task list
func (c *sopsApi) GetTaskList(ctx context.Context, header http.Header, bkBizID int64,
	keyword string) ([]*GetTaskListRst, error) {

	subPath := "/get_task_list/%d"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(GetTaskListResp)
	err := c.client.Get().
		WithContext(ctx).
		SubResourcef(subPath, bkBizID).
		WithHeaders(header).
		WithParam("keyword", keyword).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("failed to get task list by BkBizId, keyword: %s, bkBizID: %d, err: %v", keyword, bkBizID, err)
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("failed to parse get task list api, bkBizID: %d, errCode: %d, errMsg: %s",
			bkBizID, resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// sopsApi sops api interface implementation
type sopsApi struct {
	client rest.ClientInterface
	opts   cc.SopsCli
}

func (c *sopsApi) getAuthHeader() (string, string) {
	key := "X-Bkapi-Authorization"
	val := fmt.Sprintf("{\"bk_app_code\": \"%s\", \"bk_app_secret\": \"%s\", \"bk_username\":\"%s\"}", c.opts.AppCode,
		c.opts.AppSecret, c.opts.Operator)

	return key, val
}

// CreateTask create sops task
func (c *sopsApi) CreateTask(ctx context.Context, header http.Header, templateID, bkBizID int64, req *CreateTaskReq) (
	*CreateTaskResp, error) {

	subPath := "/create_task/%d/%d"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(CreateTaskResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath, templateID, bkBizID).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("failed to send create task api, templateID: %d, bkBizID: %d, err: %v, req: %+v",
			templateID, bkBizID, err, req)
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("failed to parse create task api, templateID: %d, bkBizID: %d, errCode: %d, "+
			"errMsg: %s, req: %+v", templateID, bkBizID, resp.Code, resp.Message, req)
	}

	return resp, nil
}

// StartTask start sops task
func (c *sopsApi) StartTask(ctx context.Context, header http.Header, taskID, bkBizID int64) (*StartTaskResp, error) {
	subPath := "/start_task/%d/%d"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(StartTaskResp)
	err := c.client.Post().
		WithContext(ctx).
		Body(map[string]string{}).
		SubResourcef(subPath, taskID, bkBizID).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("failed to send start task api, taskID: %d, bkBizID: %d, err: %v", taskID, bkBizID, err)
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("failed to parse start task api, taskID: %d, bkBizID: %d, errCode: %d, errMsg: %s",
			taskID, bkBizID, resp.Code, resp.Message)
	}

	return resp, nil
}

// GetTaskStatus get sops task status
func (c *sopsApi) GetTaskStatus(ctx context.Context, header http.Header, taskID, bkBizID int64) (
	*GetTaskStatusResp, error) {

	subPath := "/get_task_status/%d/%d"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(GetTaskStatusResp)
	err := c.client.Get().
		WithContext(ctx).
		SubResourcef(subPath, taskID, bkBizID).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("failed to send get task status api, taskID: %d, bkBizID: %d, err: %v", taskID, bkBizID, err)
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("failed to parse get task status api, taskID: %d, bkBizID: %d, errCode: %d, errMsg: %s",
			taskID, bkBizID, resp.Code, resp.Message)
	}

	return resp, nil
}

// GetTaskNodeDetail get sops task node detail
func (c *sopsApi) GetTaskNodeDetail(ctx context.Context, header http.Header, taskID, bkBizID int64, nodeID string) (
	*GetTaskNodeDetailResp, error) {

	subPath := "/get_task_node_detail/%d/%d"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(GetTaskNodeDetailResp)
	err := c.client.Get().
		WithContext(ctx).
		SubResourcef(subPath, taskID, bkBizID).
		WithHeaders(header).
		WithParam("node_id", nodeID).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("failed to send get task node detail api, taskID: %d, bkBizID: %d, nodeID: %s, err: %v",
			taskID, bkBizID, nodeID, err)
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("failed to parse get task node detail api, taskID: %d, bkBizID: %d, nodeID: %s, "+
			"errCode: %d, errMsg: %s", taskID, bkBizID, nodeID, resp.Code, resp.Message)
	}

	return resp, nil
}

// GetTaskNodeData get sops task node data
func (c *sopsApi) GetTaskNodeData(kt *kit.Kit, header http.Header, taskID, bkBizID int64, nodeID string,
	subStacks string) (*GetTaskNodeDataRst, error) {

	subPath := "/get_task_node_data/%d/%d"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(GetTaskNodeDataResp)
	cli := c.client.Get().
		WithContext(kt.Ctx).
		SubResourcef(subPath, bkBizID, taskID).
		WithHeaders(header).
		WithParam("node_id", nodeID)
	if len(subStacks) > 0 {
		subStacksJSON, err := json.Marshal([]string{subStacks})
		if err != nil {
			return nil, err
		}
		cli = cli.WithParam("subprocess_stack", string(subStacksJSON))
	}
	err := cli.Do().Into(resp)
	if err != nil {
		logs.Errorf("failed to send get task node data api, taskID: %d, bkBizID: %d, nodeID: %s, subStacks: %v, "+
			"err: %v, rid: %s", taskID, bkBizID, nodeID, subStacks, err, kt.Rid)
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("failed to parse get task node data api, taskID: %d, bkBizID: %d, nodeID: %s, "+
			"subStacks: %s, errCode: %d, errMsg: %s", taskID, bkBizID, nodeID, subStacks, resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// GetTaskDetail get sops task detail
func (c *sopsApi) GetTaskDetail(kt *kit.Kit, header http.Header, taskID, bkBizID int64) (
	*GetTaskDetailDataResp, error) {

	subPath := "/get_task_detail/%d/%d"
	key, val := c.getAuthHeader()
	if header == nil {
		header = http.Header{}
	}
	header.Set(key, val)
	resp := new(GetTaskDetailDataResp)
	err := c.client.Get().
		WithContext(kt.Ctx).
		SubResourcef(subPath, taskID, bkBizID).
		WithHeaders(header).
		Do().
		Into(resp)

	if err != nil {
		logs.Errorf("failed to send get task detail data api, taskID: %d, bkBizID: %d, err: %v, rid: %s",
			taskID, bkBizID, err, kt.Rid)
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		logs.Errorf("failed to parse get task node data api, taskID: %d, bkBizID: %d, errCode: %d, errMsg: %s, rid: %s",
			taskID, bkBizID, resp.Code, resp.Message, kt.Rid)
		return nil, fmt.Errorf("failed to parse get task node data api, taskID: %d, bkBizID: %d, "+
			"errCode: %d, errMsg: %s", taskID, bkBizID, resp.Code, resp.Message)
	}

	return resp, nil
}
