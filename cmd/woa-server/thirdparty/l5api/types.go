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

// Package l5api l5 api
package l5api

import (
	"fmt"
	"sync"
)

// ClientOptions client options
type ClientOptions struct {
	// l5 api address
	L5ApiAddr []string
}

// ServerDiscovery server discovery
type ServerDiscovery struct {
	name    string
	servers []string
	index   int
	sync.RWMutex
}

// GetServers return server instance address
func (s *ServerDiscovery) GetServers() ([]string, error) {
	if s == nil {
		return []string{}, nil
	}
	s.RLock()
	defer s.RUnlock()

	num := len(s.servers)
	if num == 0 {
		return []string{}, fmt.Errorf("oops, there is no %s server can be used", s.name)
	}

	if s.index < num-1 {
		s.index = s.index + 1
		return append(s.servers[s.index-1:], s.servers[:s.index-1]...), nil
	} else {
		s.index = 0
		return append(s.servers[num-1:], s.servers[:num-1]...), nil
	}
}

// GetServersChan the channel from which latest address can be reached
func (s *ServerDiscovery) GetServersChan() chan []string {
	return nil
}
