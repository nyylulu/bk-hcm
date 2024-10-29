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

// Package types ...
package types

import (
	"fmt"
)

// zk path
const (
	CR_SERV_BASEPATH        = "/cr/endpoints"
	CR_SERVCONF_BASEPATH    = "/cr/config"
	CR_SERVERROR_BASEPATH   = "/cr/errors"
	CR_SERVLANG_BASEPATH    = "/cr/language"
	CR_SERVNOTICE_BASEPATH  = "/cr/event"
	CR_SERVLIMITER_BASEPATH = "/cr/limiter"

	CR_DISCOVERY_PREFIX = "cr_"
)

// cr modules
const (
	CR_MODULE_APISERVER        = "apiserver"
	CR_MODULE_TASKSERVER       = "taskserver"
	CR_MODULE_CVMSERVER        = "cvmserver"
	CR_MODULE_CONFIGSERVER     = "configserver"
	CR_MODULE_PREDICTIONSERVER = "predictionserver"
	CR_MODULE_POOLSERVER       = "poolserver"
)

// AllModule all cr module
var AllModule = map[string]bool{
	CR_MODULE_APISERVER:        true,
	CR_MODULE_TASKSERVER:       true,
	CR_MODULE_CVMSERVER:        true,
	CR_MODULE_CONFIGSERVER:     true,
	CR_MODULE_PREDICTIONSERVER: true,
	CR_MODULE_POOLSERVER:       true,
}

// cc functionality define
const (
	CCFunctionalityServicediscover = "servicediscover"
	CCFunctionalityMongo           = "mongo"
	CCFunctionalityRedis           = "redis"
)

// ServerInfo define base server information
type ServerInfo struct {
	IP         string `json:"ip"`
	Port       uint   `json:"port"`
	RegisterIP string `json:"registerip"`
	HostName   string `json:"hostname"`
	Scheme     string `json:"scheme"`
	Version    string `json:"version"`
	Pid        int    `json:"pid"`
	// UUID is used to distinguish which service is master in zookeeper
	UUID string `json:"uuid"`
}

// APIServerServInfo apiserver information
type APIServerServInfo struct {
	ServerInfo
}

// RegisterAddress convert struct to host address
func (s *ServerInfo) RegisterAddress() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("%s://%s:%d", s.Scheme, s.RegisterIP, s.Port)
}

// Instance convert struct to host address
func (s *ServerInfo) Instance() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", s.IP, s.Port)
}
