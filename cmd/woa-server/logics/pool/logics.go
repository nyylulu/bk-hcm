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

// Package pool logics provides service logics interface
package pool

import (
	"context"

	"hcm/cmd/woa-server/thirdparty"
	"hcm/cmd/woa-server/thirdparty/esb"
	"hcm/pkg/cc"
)

// Logics provides management interface for operations of resource pool
type Logics interface {
	Pool() PoolIf
}

type logics struct {
	pool PoolIf
}

// New create a logics manager
func New(ctx context.Context, cliConf cc.ClientConfig, thirdCli *thirdparty.Client, esbCli esb.Client) Logics {
	return &logics{
		pool: NewPoolIf(ctx, cliConf, thirdCli, esbCli),
	}

}

// Pool pool interface
func (l *logics) Pool() PoolIf {
	return l.pool
}
