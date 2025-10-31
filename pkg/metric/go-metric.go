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

package metric

import (
	"fmt"
	"runtime"
)

func newGoMetricCollector() *Collector {
	golang := &golang{
		goRoutineMetric: goMetric{
			Name:    "go_goroutines",
			Help:    "Number of goroutines that currently exist.",
			GetFunc: func() float64 { return float64(runtime.NumGoroutine()) },
		},
		goProcessMetric: goMetric{
			Name: "go_threads",
			Help: "Number of OS threads created",
			GetFunc: func() float64 {
				n, _ := runtime.ThreadCreateProfile(nil)
				return float64(n)
			},
		},
		goCPUMetric: goMetric{
			Name:    "go_cpu_used",
			Help:    " the number of logical CPUs usable by the current process.",
			GetFunc: func() float64 { return float64(runtime.NumCPU()) },
		},
		goMemStateMetrics: []goMetric{
			{
				Name:       memstatNamespace("alloc_bytes"),
				Help:       "Number of bytes allocated and still in use.",
				MemGetFunc: func(ms *runtime.MemStats) float64 { return float64(ms.Alloc) },
			},
			{
				Name:       memstatNamespace("alloc_bytes_total"),
				Help:       "Total number of bytes allocated, even if freed.",
				MemGetFunc: func(ms *runtime.MemStats) float64 { return float64(ms.TotalAlloc) },
			},
			{
				Name:       memstatNamespace("sys_bytes"),
				Help:       "Number of bytes obtained from system.",
				MemGetFunc: func(ms *runtime.MemStats) float64 { return float64(ms.Sys) },
			},
			{
				Name:       memstatNamespace("mallocs_total"),
				Help:       "Total number of mallocs.",
				MemGetFunc: func(ms *runtime.MemStats) float64 { return float64(ms.Mallocs) },
			},
			{
				Name:       memstatNamespace("frees_total"),
				Help:       "Total number of frees.",
				MemGetFunc: func(ms *runtime.MemStats) float64 { return float64(ms.Frees) },
			},
		},
	}

	return NewCollector("golang_metrics", golang)
}

func memstatNamespace(s string) string {
	return fmt.Sprintf("go_memstats_%s", s)
}

type golang struct {
	goRoutineMetric   goMetric
	goProcessMetric   goMetric
	goCPUMetric       goMetric
	goMemStateMetrics []goMetric
}

// Collect ...
func (g golang) Collect() []MetricInterf {
	m := make([]MetricInterf, 0)
	m = append(m, g.goRoutineMetric)
	m = append(m, g.goProcessMetric)

	ms := &runtime.MemStats{}
	runtime.ReadMemStats(ms)
	for idx := range g.goMemStateMetrics {
		g.goMemStateMetrics[idx].MemStats = ms
		m = append(m, &g.goMemStateMetrics[idx])
	}
	return m
}

type goMetric struct {
	Name       string
	Help       string
	MemStats   *runtime.MemStats
	MemGetFunc func(stat *runtime.MemStats) float64
	GetFunc    func() float64
}

// GetMeta ...
func (m goMetric) GetMeta() *MetricMeta {
	return &MetricMeta{
		Name: m.Name,
		Help: m.Help,
	}
}

// GetValue ..
func (m goMetric) GetValue() (*FloatOrString, error) {
	if m.MemStats != nil {
		return FormFloatOrString(m.MemGetFunc(m.MemStats))
	}
	return FormFloatOrString(m.GetFunc())
}

// GetExtension ...
func (m goMetric) GetExtension() (*MetricExtension, error) {
	return nil, nil
}
