/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package dispatcher

import (
	"hcm/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

// dispatcherMetrics is used to collect recycle dispatcher metrics.
var dispatcherMetrics *metric

// InitDispatcherMetrics ...
func InitDispatcherMetrics(reg prometheus.Registerer) {
	m := new(metric)

	m.OrderStateCostSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Subsystem: metrics.HostRecycleSubSys,
		Name:      "order_state_cost_seconds",
		Help:      "the cost seconds of specific recycle order state",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 3, 5, 10, 20, 30, 45, 90,
			120, 180, 300, 600, 1800, 3600, 7200, 10800, 21600, 43200, 86400},
	}, []string{"status", "bk_biz_id"})
	reg.MustRegister(m.OrderStateCostSec)

	m.OrderStateErrCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: metrics.HostRecycleSubSys,
			Name:      "order_state_err_count",
			Help:      "the total error count of specific recycle order of state",
		}, []string{"status", "bk_biz_id"})
	reg.MustRegister(m.OrderStateErrCounter)

	//  回收单据提交到当前状态耗时
	m.RecycleStateCostSinceCommitSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Subsystem: metrics.HostRecycleSubSys,
		Name:      "order_state_cost_since_commit_seconds",
		Help:      "the cost seconds from submit to specific recycle order state",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 3, 5, 10, 20, 30, 45, 90, 120, 180, 300, 600, 1800, 3600, 7200,
			10800, 21600, 43200, 86400},
	}, []string{"status", "bk_biz_id"})
	reg.MustRegister(m.RecycleStateCostSinceCommitSec)

	dispatcherMetrics = m
}

type metric struct {

	// 回收状态流转耗时
	OrderStateCostSec *prometheus.HistogramVec
	// 回收状态流转错误计数器
	OrderStateErrCounter *prometheus.CounterVec
	// 提交单据到当前状态耗时
	RecycleStateCostSinceCommitSec *prometheus.HistogramVec
}
