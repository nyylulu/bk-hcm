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

package detector

import (
	"hcm/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	labelStepName  = "step_name"
	labelSopsState = "state"
)

// detectorMetrics is used to collect detector metrics.
var detectorMetrics *metric

// InitDetectorMetrics ...
func InitDetectorMetrics(reg prometheus.Registerer) {
	m := new(metric)

	m.DetectStepCostSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Subsystem: metrics.HostRecycleSubSys,
		Name:      "detect_step_cost_seconds",
		Help:      "the cost seconds to specific detect step",
		Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.85, 1, 2, 3, 4, 5, 10, 12.5, 15, 17.5, 20, 22.5, 25, 30,
			35, 40, 45, 50, 55, 60, 65, 75, 90, 120, 180, 300, 600, 1800, 3600},
	}, []string{labelStepName})
	reg.MustRegister(m.DetectStepCostSec)

	m.DetectStepErrCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metrics.Namespace,
			Subsystem: metrics.HostRecycleSubSys,
			Name:      "detect_step_err_count",
			Help:      "the total error count to specific detect step",
		}, []string{labelStepName})
	reg.MustRegister(m.DetectStepErrCounter)

	m.SopsStepCostSec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metrics.Namespace,
		Subsystem: metrics.HostRecycleSubSys,
		Name:      "sops_phase_cost_seconds",
		Help:      "the cost seconds to sops",
		Buckets: []float64{0.5, 1, 2.5, 5, 7.5, 10, 12.5, 15, 17.5, 20, 21, 22, 23, 24, 25, 27.5, 30,
			35, 40, 45, 50, 55, 60, 65, 75, 90, 120, 180, 300, 600, 1800, 3600},
	}, []string{labelSopsState})
	reg.MustRegister(m.SopsStepCostSec)

	detectorMetrics = m
}

type metric struct {

	// 回收预检当前步骤耗时
	DetectStepCostSec *prometheus.HistogramVec
	// 回收预检步骤错误计数器
	DetectStepErrCounter *prometheus.CounterVec
	// 标准运维预检不同阶段（启动、查询到结果）耗时
	SopsStepCostSec *prometheus.HistogramVec
}
