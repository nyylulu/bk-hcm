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

// Package utils provides common utils
package utils

import (
	"fmt"
	"runtime"
	"time"

	"hcm/pkg/logs"
)

// Retry redoes op until success or timeout
func Retry(op func() (interface{}, error), checker func(interface{}, error) (bool, error), timeout uint64,
	interval uint64) (ret interface{}, err error) {
	pc, _, _, _ := runtime.Caller(1)
	callerName := runtime.FuncForPC(pc).Name()

	var tm <-chan time.Time
	tm = time.After(time.Duration(timeout) * time.Second)

	// need to reassign err and isTimeout value
	times := 0
	for {
		times = times + 1
		select {
		case <-tm:
			err = fmt.Errorf("Retry (#%d) %s timeout!", times, callerName)
			return
		default:
		}
		ret, err = op()
		if ok, message := checker(ret, err); ok {
			err = message
			logs.Infof("Retry (#%d) %s succ, message: %v", times, callerName, message)
			return
		} else {
			// retry again if checker return false
			logs.Infof("Retry (#%d) %s again, message: %v", times, callerName, message)
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
