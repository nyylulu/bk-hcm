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

// Package util provides utility functions
package util

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/emicklei/go-restful"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"

	"hcm/pkg"
	"hcm/pkg/criteria/errors"
)

// InStrArr check if key is in arr
func InStrArr(arr []string, key string) bool {
	for _, a := range arr {
		if key == a {
			return true
		}
	}
	return false
}

// GetLanguage get language from header
func GetLanguage(header http.Header) string {
	return header.Get(pkg.BKHTTPLanguage)
}

// GetUser get user from header
func GetUser(header http.Header) string {
	return header.Get(pkg.BKHTTPHeaderUser)
}

// GetOwnerID get user from header
func GetOwnerID(header http.Header) string {
	return header.Get(pkg.BKHTTPOwnerID)
}

// SetOwnerIDAndAccount set supplier id and account in head
func SetOwnerIDAndAccount(req *restful.Request) {
	owner := req.Request.Header.Get(pkg.BKHTTPOwner)
	if "" != owner {
		req.Request.Header.Set(pkg.BKHTTPOwnerID, owner)
	}
}

// GetHTTPCCRequestID return config center request id from http header
func GetHTTPCCRequestID(header http.Header) string {
	rid := header.Get(pkg.BKHTTPCCRequestID)
	return rid
}

// ExtractRequestIDFromContext extract request id from context
func ExtractRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	rid := ctx.Value(pkg.ContextRequestIDField)
	ridValue, ok := rid.(string)
	if ok == true {
		return ridValue
	}
	return ""
}

// ExtractOwnerFromContext extract supplier id from context
func ExtractOwnerFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	owner := ctx.Value(pkg.ContextRequestOwnerField)
	ownerValue, ok := owner.(string)
	if ok == true {
		return ownerValue
	}
	return ""
}

// NewContextFromGinContext create a new context from gin context
func NewContextFromGinContext(c *gin.Context) context.Context {
	return NewContextFromHTTPHeader(c.Request.Header)
}

// NewContextFromHTTPHeader create a new context from http header
func NewContextFromHTTPHeader(header http.Header) context.Context {
	rid := GetHTTPCCRequestID(header)
	user := GetUser(header)
	owner := GetOwnerID(header)
	ctx := context.Background()
	ctx = context.WithValue(ctx, pkg.ContextRequestIDField, rid)
	ctx = context.WithValue(ctx, pkg.ContextRequestUserField, user)
	ctx = context.WithValue(ctx, pkg.ContextRequestOwnerField, owner)
	return ctx
}

// NewHeaderFromContext create a new header from context
func NewHeaderFromContext(ctx context.Context) http.Header {
	rid := ctx.Value(pkg.ContextRequestIDField)
	ridValue, ok := rid.(string)
	if !ok {
		ridValue = GenerateRID()
	}

	user := ctx.Value(pkg.ContextRequestUserField)
	userValue, ok := user.(string)
	if !ok {
		ridValue = "admin"
	}

	owner := ctx.Value(pkg.ContextRequestOwnerField)
	ownerValue, ok := owner.(string)
	if !ok {
		ownerValue = pkg.BKDefaultOwnerID
	}

	header := make(http.Header)
	header.Set(pkg.BKHTTPCCRequestID, ridValue)
	header.Set(pkg.BKHTTPHeaderUser, userValue)
	header.Set(pkg.BKHTTPOwnerID, ownerValue)

	header.Add("Content-Type", "application/json")

	return header
}

// BuildHeader build a header from user and supplier account
func BuildHeader(user string, supplierAccount string) http.Header {
	header := make(http.Header)
	header.Add(pkg.BKHTTPOwnerID, supplierAccount)
	header.Add(pkg.BKHTTPHeaderUser, user)
	header.Add(pkg.BKHTTPCCRequestID, GenerateRID())
	header.Add("Content-Type", "application/json")
	return header
}

// ExtractRequestUserFromContext extract user from context
func ExtractRequestUserFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	user := ctx.Value(pkg.ContextRequestUserField)
	userValue, ok := user.(string)
	if ok == true {
		return userValue
	}
	return ""
}

// AtomicBool is an atomic bool
type AtomicBool int32

// NewBool create a new AtomicBool
func NewBool(yes bool) *AtomicBool {
	var n = AtomicBool(0)
	if yes {
		n = AtomicBool(1)
	}
	return &n
}

// SetIfNotSet set the value to 1 if it was not set before
func (b *AtomicBool) SetIfNotSet() bool {
	return atomic.CompareAndSwapInt32((*int32)(b), 0, 1)
}

// Set set the value
func (b *AtomicBool) Set() {
	atomic.StoreInt32((*int32)(b), 1)
}

// UnSet set the value to 0
func (b *AtomicBool) UnSet() {
	atomic.StoreInt32((*int32)(b), 0)
}

// IsSet check if the value is set
func (b *AtomicBool) IsSet() bool {
	return atomic.LoadInt32((*int32)(b)) == 1
}

// SetTo set the value to 1 or 0
func (b *AtomicBool) SetTo(yes bool) {
	if yes {
		atomic.StoreInt32((*int32)(b), 1)
	} else {
		atomic.StoreInt32((*int32)(b), 0)
	}
}

// IntSlice ...
type IntSlice []int

// Len ...
func (p IntSlice) Len() int { return len(p) }

// Less ...
func (p IntSlice) Less(i, j int) bool { return p[i] < p[j] }

// Swap ...
func (p IntSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// Int64Slice ...
type Int64Slice []int64

// Len ...
func (p Int64Slice) Len() int { return len(p) }

// Less ...
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }

// Swap ...
func (p Int64Slice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// GenerateRID ...
func GenerateRID() string {
	unused := "0000"
	id := xid.New()
	return fmt.Sprintf("cc%s%s", unused, id.String())
}

// Int64Join []int64 to string
func Int64Join(data []int64, separator string) string {
	var ret string
	for _, item := range data {
		ret += strconv.FormatInt(item, 10) + separator
	}
	return strings.Trim(ret, separator)
}

// BuildMongoField build mongodb sub item field key
func BuildMongoField(key ...string) string {
	return strings.Join(key, ".")
}

// BuildMongoSyncItemField build mongodb sub item synchronize field key
func BuildMongoSyncItemField(key string) string {
	return BuildMongoField(pkg.MetadataField, pkg.MetaDataSynchronizeField, key)
}

// GetDefaultCCError get default CCErrorIf
func GetDefaultCCError(header http.Header) errors.DefaultCCErrorIf {
	globalCCError := errors.GetGlobalCCError()
	if globalCCError == nil {
		return nil
	}
	language := GetLanguage(header)
	return globalCCError.CreateDefaultCCErrorIf(language)
}

// CCHeader get cc header
func CCHeader(header http.Header) http.Header {
	newHeader := make(http.Header, 0)
	newHeader.Add(pkg.BKHTTPCCRequestID, header.Get(pkg.BKHTTPCCRequestID))
	newHeader.Add(pkg.BKHTTPCookieLanugageKey, header.Get(pkg.BKHTTPCookieLanugageKey))
	newHeader.Add(pkg.BKHTTPHeaderUser, header.Get(pkg.BKHTTPHeaderUser))
	newHeader.Add(pkg.BKHTTPLanguage, header.Get(pkg.BKHTTPLanguage))
	newHeader.Add(pkg.BKHTTPOwner, header.Get(pkg.BKHTTPOwner))
	newHeader.Add(pkg.BKHTTPOwnerID, header.Get(pkg.BKHTTPOwnerID))
	newHeader.Add(pkg.BKHTTPRequestAppCode, header.Get(pkg.BKHTTPRequestAppCode))
	newHeader.Add(pkg.BKHTTPRequestRealIP, header.Get(pkg.BKHTTPRequestRealIP))
	newHeader.Add(pkg.BKHTTPReadReference, header.Get(pkg.BKHTTPReadReference))

	return newHeader
}

// SetHTTPReadPreference  再header 头中设置mongodb read preference， 这个是给调用子流程使用
func SetHTTPReadPreference(header http.Header, mode pkg.ReadPreferenceMode) http.Header {
	header.Set(pkg.BKHTTPReadReference, mode.String())
	return header
}

// SetDBReadPreference  再context 设置设置mongodb read preference，给dal 使用
func SetDBReadPreference(ctx context.Context, mode pkg.ReadPreferenceMode) context.Context {
	ctx = context.WithValue(ctx, pkg.BKHTTPReadReference, mode.String())
	return ctx
}

// SetReadPreference  再context， header 设置设置mongodb read preference，给dal 使用
func SetReadPreference(ctx context.Context, header http.Header, mode pkg.ReadPreferenceMode) (context.Context,
	http.Header) {

	ctx = SetDBReadPreference(ctx, mode)
	header = SetHTTPReadPreference(header, mode)
	return ctx, header
}

// GetDBReadPreference 从context中获取mongodb read preference
func GetDBReadPreference(ctx context.Context) pkg.ReadPreferenceMode {
	val := ctx.Value(pkg.BKHTTPReadReference)
	if val != nil {
		mode, ok := val.(string)
		if ok {
			return pkg.ReadPreferenceMode(mode)
		}
	}
	return pkg.NilMode
}

// GetHTTPReadPreference ...
func GetHTTPReadPreference(header http.Header) pkg.ReadPreferenceMode {
	mode := header.Get(pkg.BKHTTPReadReference)
	if mode == "" {
		return pkg.NilMode
	}
	return pkg.ReadPreferenceMode(mode)
}

// GetGatewayName get BKHTTPGatewayName from header
func GetGatewayName(header http.Header) string {
	return header.Get(pkg.BKHTTPGatewayName)
}

// GetJWTToken get jwt token from header
func GetJWTToken(header http.Header) string {
	return header.Get(pkg.BKHTTPJWTToken)
}

// ValidHeaderFromAPIGW valid header from apigateway
func ValidHeaderFromAPIGW(header http.Header) (bool, string) {
	return true, ""
}

// SetHeaderFromApiGW set header from apigateway
func SetHeaderFromApiGW(header http.Header, userName, appCode string) {
	header.Set(pkg.BKHTTPLanguage, header.Get(pkg.BKHTTPAPIGWLanguage))
	header.Set(pkg.BKHTTPOwnerID, header.Get(pkg.BKHTTPAPIGWOwnerID))
	header.Set(pkg.BKHTTPHeaderUser, userName)
	header.Set(pkg.BKHTTPRequestAppCode, appCode)
}
