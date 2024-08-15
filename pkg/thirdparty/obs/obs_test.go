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

package obs

import (
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"hcm/pkg/cc"
	"hcm/pkg/kit"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestIEGObs_NotifyRePull(t *testing.T) {
	var temporalAddr = "127.0.0.1:56789"
	kt := kit.New()
	type args struct {
		kt  *kit.Kit
		req *NotifyObsPullReq
	}

	opt := &cc.IEGObsOption{
		Endpoints: []string{
			"http://" + temporalAddr,
		},
		APIKey: "abcdefg",
		TLS:    cc.TLSConfig{},
	}

	dec, _ := decimal.NewFromString("567891234.0123456789")

	tests := []struct {
		name       string
		args       args
		wantErr    bool
		path       string
		wantedBody string
		resp       string
	}{
		{
			name: "test repull aws",
			args: args{
				kt: kt,
				req: &NotifyObsPullReq{
					YearMonth: 202406,
					AccountInfoList: []AccountInfo{{
						AccountType: AccountTypeAws,
						Total:       1234,
						Column:      "cost",
						SumColValue: dec,
					}},
				},
			},
			wantErr: false,
			path:    "/jsonrpc/obs-api?api_key=abcdefg",
			wantedBody: `{
    "jsonrpc": "2.0",
    "id": "0",
    "method": "rePullIegAccountData",
    "params": {
        "yearMonth": 202406,
        "accountInfoList": "[{\"accountType\": \"AWS\",\"total\": 1234,\"column\": \"cost\",\"sumColValue\": 567891234.0123456789}]"
    }
}`,
			resp: `{
    "jsonrpc": "2.0",
    "id": "0",
    "result": {
        "data": "OK",
        "message": "成功",
        "status": 0
    }
}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obsCli, err := NewIEGObs(opt, nil)
			if err != nil {
				t.Fatalf("fail to create obs client, err: %v", err)
			}
			// Start a local HTTP server
			mockServer := http.Server{Addr: temporalAddr,
				Handler: getHandler(t, tt.path, tt.wantedBody, tt.resp)}

			go func() {
				err := mockServer.ListenAndServe()
				if !errors.Is(err, http.ErrServerClosed) {
					t.Errorf("fail to start mock server, err: %v", err)
					return
				}
			}()
			defer mockServer.Close()
			time.Sleep(time.Second)
			if err := obsCli.NotifyRePull(tt.args.kt, tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("NotifyRePull() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func getHandler(t *testing.T, path string, wantedBody, resp string) http.HandlerFunc {

	return func(rw http.ResponseWriter, req *http.Request) {

		assert.Equal(t, path, req.URL.Path)

		buf, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("fail to read request: %v", err)
		}
		if len(buf) == 0 {
			t.Logf("no request body")
		}

		assert.JSONEq(t, wantedBody, string(buf))
		_, err = rw.Write([]byte(resp))
		if err != nil {
			t.Errorf("fail to write response, %v", err)
		}
	}
}

func TestNotifyObsPullReq_Validate(t *testing.T) {

	tests := []struct {
		name    string
		arg     NotifyObsPullReq
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "empty",
			arg: NotifyObsPullReq{
				YearMonth:       0,
				AccountInfoList: nil,
			},
			wantErr: func(t assert.TestingT, err error, msg ...interface{}) bool {

				assert.NotNil(t, err)

				var e validator.ValidationErrors
				ok := errors.As(err, &e)
				assert.Equal(t, true, ok)
				assert.Len(t, e, 2)
				assert.Equal(t, "yearMonth", e[0].Field(), msg...)
				assert.Equal(t, "required", e[0].Tag(), msg...)

				assert.Equal(t, "accountInfoList", e[1].Field(), msg...)
				assert.Equal(t, "required", e[1].Tag(), msg...)
				return false
			},
		},
		{
			name: "empty accountInfoEmpty",
			arg: NotifyObsPullReq{
				YearMonth: 202406,
				AccountInfoList: []AccountInfo{
					{
						AccountType: "",
						Total:       0,
						Column:      "",
						SumColValue: decimal.Decimal{},
					},
				},
			},
			wantErr: func(t assert.TestingT, err error, msg ...interface{}) bool {
				assert.NotNil(t, err)
				var e validator.ValidationErrors
				ok := errors.As(err, &e)
				assert.Equal(t, true, ok)
				assert.Len(t, e, 3)

				assert.Equal(t, "accountType", e[0].Field(), msg...)
				assert.Equal(t, "required", e[0].Tag(), msg...)

				assert.Equal(t, "total", e[1].Field(), msg...)
				assert.Equal(t, "required", e[1].Tag(), msg...)

				assert.Equal(t, "column", e[2].Field(), msg...)
				assert.Equal(t, "required", e[2].Tag(), msg...)
				return true
			},
		},
		{
			name: "all match",
			arg: NotifyObsPullReq{
				YearMonth: 202406,
				AccountInfoList: []AccountInfo{
					{
						AccountType: AccountTypeGCP,
						Total:       10,
						Column:      "cost",
						SumColValue: decimal.Decimal{},
					},
				},
			},
			wantErr: func(t assert.TestingT, err error, msg ...interface{}) bool {
				assert.Nil(t, err)
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &tt.arg
			tt.wantErr(t, r.Validate(), tt.name)
		})
	}
}

func TestNotifyObsPullArg(t *testing.T) {
	var temporalAddr = "127.0.0.1:56789"
	kt := kit.New()

	opt := &cc.IEGObsOption{
		Endpoints: []string{
			"http://" + temporalAddr,
		},
		APIKey: "abcdefg",
		TLS:    cc.TLSConfig{},
	}
	tests := []struct {
		name    string
		arg     *NotifyObsPullReq
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "nil",
			arg:  nil,
			wantErr: func(t assert.TestingT, err error, msg ...interface{}) bool {
				assert.EqualError(t, err, "NotifyObsPullReq is required")
				return false
			},
		},
		{
			name: "empty",
			arg: &NotifyObsPullReq{
				YearMonth:       0,
				AccountInfoList: nil,
			},
			wantErr: func(t assert.TestingT, err error, msg ...interface{}) bool {

				assert.NotNil(t, err)

				var e validator.ValidationErrors
				ok := errors.As(err, &e)
				assert.Equal(t, true, ok)
				assert.Len(t, e, 2)
				assert.Equal(t, "yearMonth", e[0].Field(), msg...)
				assert.Equal(t, "required", e[0].Tag(), msg...)

				assert.Equal(t, "accountInfoList", e[1].Field(), msg...)
				assert.Equal(t, "required", e[1].Tag(), msg...)
				return false
			},
		},
		{
			name: "empty accountInfoEmpty",
			arg: &NotifyObsPullReq{
				YearMonth: 202406,
				AccountInfoList: []AccountInfo{
					{
						AccountType: "",
						Total:       0,
						Column:      "",
						SumColValue: decimal.Decimal{},
					},
				},
			},
			wantErr: func(t assert.TestingT, err error, msg ...interface{}) bool {
				assert.NotNil(t, err)
				var e validator.ValidationErrors
				ok := errors.As(err, &e)
				assert.Equal(t, true, ok)
				assert.Len(t, e, 3)

				assert.Equal(t, "accountType", e[0].Field(), msg...)
				assert.Equal(t, "required", e[0].Tag(), msg...)

				assert.Equal(t, "total", e[1].Field(), msg...)
				assert.Equal(t, "required", e[1].Tag(), msg...)

				assert.Equal(t, "column", e[2].Field(), msg...)
				assert.Equal(t, "required", e[2].Tag(), msg...)
				return true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obsCli, err := NewIEGObs(opt, nil)
			if err != nil {
				t.Fatalf("fail to create obs client, err: %v", err)
			}
			tt.wantErr(t, obsCli.NotifyRePull(kt, tt.arg), tt.name)
		})
	}
}
