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
	"encoding/json"
	"time"

	"hcm/pkg/logs"
)

var metricController *MetricController

// MetricController ...
type MetricController struct {
	MetaData   *MetaData
	Collectors map[CollectorName]CollectInter
}

// PackMetrics ...
func (mc *MetricController) PackMetrics() (*[]byte, error) {
	mf := MetricFamily{
		MetaData:     mc.MetaData,
		MetricBundle: make(map[CollectorName][]*Metric),
	}

	for name, collector := range mc.Collectors {
		mf.MetricBundle[name] = make([]*Metric, 0)
		done := make(chan struct{}, 0)
		go func(c CollectInter) {
			for _, mc := range c.Collect() {
				metric, err := newMetric(mc)
				if nil != err {
					logs.Errorf("new metric failed. err: %v", err)
					continue
				}
				mf.MetricBundle[name] = append(mf.MetricBundle[name], metric)
			}
			done <- struct{}{}
		}(collector)

		select {
		case <-time.After(time.Duration(10 * time.Second)):
			logs.Errorf("get metric bundle: %s timeout, skip it.", name)
			continue
		case <-done:
			close(done)
		}
	}

	mf.ReportTimeMs = time.Now().Unix()
	js, err := json.Marshal(mf)
	if nil != err {
		return nil, err
	}
	return &js, nil
}
