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

package util

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetCurrentTimeStr(t *testing.T) {
	now := time.Now()
	val := GetCurrentTimeStr()
	valTime, err := time.ParseInLocation("2006-01-02 15:04:05", val, time.Local)
	require.NoError(t, err)
	require.InDelta(t, now.Unix(), valTime.Unix(), 1)
}

func TestConvParamsTime(t *testing.T) {
	strJSON := `{"bk_host_id":{"$in":[99,100,101,102,103,104]},"create_time":{"$in":["2018-03-16 02:45:28","2018-03-16"]}}`
	var a interface{}
	err := json.Unmarshal([]byte(strJSON), &a)
	if nil != err {
		t.Error(err.Error())
	}
	fmt.Println("====================")
	a = ConvParamsTime(a)
	fmt.Println(a)

}

func TestFormatPeriod(t *testing.T) {
	period := "000290S"
	periodFormated, err := FormatPeriod(period)
	if nil != err {
		t.Error(err.Error())
		return
	}
	if periodFormated != "290S" {
		t.Errorf("error formated period %s", periodFormated)
	}
	fmt.Println(periodFormated)
}

func TestTimeStrToUnixSecondDefault(t *testing.T) {
	timeStr := "2024-09-24"
	timeUnix, err := TimeStrToUnixSecondDefault(timeStr)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if timeUnix != 1727107200 {
		t.Errorf("error time unix %d", timeUnix)
	}
	fmt.Println(timeUnix)
}

func TestTimeStrToTimePtr(t *testing.T) {
	timeStr := "2024-09-24"
	timePtr, err := TimeStrToTimePtr(timeStr)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !timePtr.Equal(time.Date(2024, 9, 24, 0, 0, 0, 0, time.Local)) {
		t.Error("error time ptr")
	}
	fmt.Println(timePtr)
	fmt.Println(timePtr.Unix())
	fmt.Println(timePtr.Format("2006-01-02 15:04:05"))
}
