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

// Package safetyapi is the client for safetyapi
package safetyapi

// response exmaple:
/*
{
    "ret":0,
    "msg":"ok",
    "data":{
        "page":1,
        "page_size":10,
        "total_count":1,
        "data":[
            {
                "id":"264156",
                "ip":"10.0.0.1",
                "container_ip":"10.0.0.2",
                "container_name":"docker-210105120530009836",
                "container_id":"4f12363914c6b32233717afd4545d113d92ca10d67e5b82ee01424d2b9d6d364",
                "path":"\/data\/esenv\/elasticsearch-7.7.0\/lib\/log4j-core-2.11.1.jar",
                "ver":"2.11.1",
                "business":"CC_reborn--CC_reborn--CC__数据待清理",
                "bg":"IEG",
                "dept":"技术运营部",
                "operator":"",
                "owneruin":"",
                "uuid":"4f12363914c6b32233717afd4545d113d92c",
                "mac":"",
                "md5":"b2242de0677be6515d6cefbf48e7e5d5",
                "scan_time":"2021-12-16 13:22:14"
            }
        ]
    }
}
*/

// BaseLineRsp is struct of security baseline response
type BaseLineRsp struct {
	Ret  int           `json:"ret"`
	Msg  string        `json:"msg"`
	Data *BaseLineData `json:"data"`
}

// BaseLineData is struct of security baseline data
type BaseLineData struct {
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalCount int          `json:"total_count"`
	Data       []*Log4jData `json:"data"`
}

// Log4jData is struct of log4j security baseline data
type Log4jData struct {
	Ip          string `json:"ip"`
	ContainerIp string `json:"container_ip"`
}
