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

// Package alarmapi is a client for Third Party API
package alarmapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/rest/client"
	cvt "hcm/pkg/tools/converter"

	hmac_auth "git.woa.com/nops/ngate/ngate-sdk/ngate-go/ngatehmac"

	"github.com/prometheus/client_golang/prometheus"
)

// AlarmClientInterface Alarm api interface
type AlarmClientInterface interface {
	// CheckAlarm returns if a host pass Alarm alarm policy check or not
	CheckAlarm(ctx context.Context, header http.Header, ip string) ([]interface{}, error)
	// AddShieldAlarm add shield alarm
	AddShieldAlarm(kt *kit.Kit, ips []string, hourNum time.Duration, operateIP string, reason string) ([]string, error)
	// DelShieldAlarm del shield alarm
	DelShieldAlarm(kt *kit.Kit, ids []string, operateIP string) ([]bool, error)
}

// NewAlarmClientInterface creates a alarm api instance
func NewAlarmClientInterface(opts cc.AlarmCli, reg prometheus.Registerer) (AlarmClientInterface, error) {
	cli, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}

	c := &client.Capability{
		Client: cli,
		Discover: &ServerDiscovery{
			name:    "alarm ngate api",
			servers: []string{opts.AlarmApiAddr},
		},
		MetricOpts: client.MetricOption{Register: reg},
	}

	return &alarmApi{
		client: rest.NewClient(c, "/"),
		opts:   opts,
	}, nil
}

// alarmApi Alarm api interface implementation
type alarmApi struct {
	client rest.ClientInterface
	opts   cc.AlarmCli
}

// CheckAlarm returns if a host pass alarm policy check or not
func (a *alarmApi) CheckAlarm(ctx context.Context, header http.Header, ip string) ([]interface{}, error) {
	return a.getShieldConfig(ctx, header, ip)
}

func (a *alarmApi) getShieldConfig(ctx context.Context, header http.Header, ip string) ([]interface{}, error) {
	req := &CheckAlarmReq{
		Method: "alarm.get_alarm_shield_config",
		Params: &CheckAlarmParams{
			Ip: ip,
		},
	}

	subPath := "/tnm2_api/alarm.get_alarm_shield_config"
	ret := make([]interface{}, 0)
	err := a.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(&ret)

	if err != nil {
		return nil, err
	}

	// 告警平台返回的响应示例：[总条目数, 起始条目索引, [详细数据{},{},...]]
	// 示例：[2,0,[
	//    {"id":63493720,"cycle":"[[\"date\",\"2022-03-28\",\"2024-03-28\",\"09:58\",\"12:58\"]]","creat_time":"2022-03-28 10:00:12","create_time":"2022-03-28 10:00:12","reason":"全屏蔽","cycle_start":"2022-03-28","cycle_end":"2024-03-28","shield_rule":"[\"true\"]","creator":"chelseacui","ciset_info":" 网络平台部->[13]ECN->[企业接入][CloudVPN]->[专线][开发测试_FCR]<br>机房:重庆腾讯泰和DC3号楼A区M3103","ciset_rule":"[[\"==\", \"_dept\", \"\\u7f51\\u7edc\\u5e73\\u53f0\\u90e8\"], [\"==\", \"_servicegroup_id\", 1111829], [\"==\", \"_service_id\", 1215101], [\"==\", \"_module_id\", 1512760], [\"==\", \"_idc\", \"\\u91cd\\u5e86\\u817e\\u8baf\\u6cf0\\u548cDC3\\u53f7\\u697cA\\u533aM3103\"]]","valid":1,"status":"屏蔽中"}
	//]]
	if len(ret) != 3 {
		return nil, fmt.Errorf("check alarm policy got invalid response format, resp: %+v", ret)
	}

	return ret[2].([]interface{}), nil
}

// addShieldConfig add shield alarm config
func (a *alarmApi) addShieldConfig(ctx context.Context, header http.Header, req *AddShieldReq) (*AddShieldResp, error) {
	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	header = a.authHeader(params, header)
	subPath := "/tnm2_api/alarmadapter.add_alarm_shield_config"
	resp := new(AddShieldResp)
	err = a.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(&resp)

	if err != nil {
		logs.Errorf("failed to add alarm shield config, err: %v, req: %+v, ips: %v",
			err, cvt.PtrToVal(req), req.Params.Ip)
		return nil, err
	}

	return resp, err
}

// delShieldConfig del shield alarm config
func (a *alarmApi) delShieldConfig(ctx context.Context, header http.Header, req *DelShieldReq) ([]bool, error) {
	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	header = a.authHeader(params, header)
	subPath := "/tnm2_api/alarm.del_alarm_shield_config_bu"
	ret := make([]bool, 0)
	err = a.client.Post().
		WithContext(ctx).
		Body(req).
		SubResourcef(subPath).
		WithHeaders(header).
		Do().
		Into(&ret)

	if err != nil {
		logs.Errorf("failed to del alarm shield config, err: %v, req: %+v, id: %s",
			err, cvt.PtrToVal(req), req.Params.ID)
		return nil, err
	}

	return ret, nil
}

func (a *alarmApi) authHeader(params []byte, header http.Header) http.Header {
	authHeader := hmac_auth.GetAuthHeader(a.opts.AppCode, a.opts.AppSecret, params)
	if header == nil {
		header = http.Header{}
	}
	for k, v := range authHeader {
		header.Set(k, v)
	}
	return header
}

// AddShieldAlarm add shield alarm
func (a *alarmApi) AddShieldAlarm(kt *kit.Kit, ips []string, hourNum time.Duration, operateIP string, reason string) (
	[]string, error) {

	shieldStart := time.Now().Format("2006-01-02 15:04")
	shieldEnd := time.Now().Add(time.Hour * hourNum).Format("2006-01-02 15:04")

	alarmIDs := make([]string, 0)
	maxNum := 100
	length := len(ips)
	for i := 0; i < length; i += maxNum {
		begin := i
		end := i + maxNum
		if end > length {
			end = length
		}

		req := &AddShieldReq{
			Method: AddShieldMethod,
			Params: &AddShieldParams{
				Ip:          ips[begin:end],
				Operator:    Operator,
				OIp:         operateIP,
				Reason:      reason,
				ShieldStart: shieldStart,
				ShieldEnd:   shieldEnd,
			},
		}

		resp, err := a.addShieldConfig(kt.Ctx, kt.Header(), req)
		if err != nil {
			// add shield config may fail, ignore it
			logs.Errorf("failed to add shield alarm config, err: %v, ips: %v, rid: %s", err, ips, kt.Rid)
			continue
		}

		if resp.Code != 0 {
			// add shield config may fail, ignore it
			logs.Errorf("failed to add shield alarm config, code: %d, msg: %s, ips: %v, rid: %s",
				resp.Code, resp.Msg, ips, kt.Rid)
			continue
		}
		alarmIDs = append(alarmIDs, strconv.FormatInt(resp.Data, 10))
	}

	return alarmIDs, nil
}

// DelShieldAlarm del shield alarm
func (a *alarmApi) DelShieldAlarm(kt *kit.Kit, ids []string, operateIP string) ([]bool, error) {
	req := &DelShieldReq{
		Method: DelShieldMethod,
		Params: &DelShieldParams{
			ID:       fmt.Sprintf("[%s]", strings.Join(ids, ",")),
			Operator: Operator,
			OIp:      operateIP,
		},
	}

	resp, err := a.delShieldConfig(kt.Ctx, kt.Header(), req)
	if err != nil {
		logs.Errorf("failed to del shield alarm config, err: %v, ids: %v, rid: %s", err, ids, kt.Rid)
		return nil, err
	}

	return resp, nil
}
