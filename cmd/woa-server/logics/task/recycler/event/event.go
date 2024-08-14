/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package event defines recycle order events
// during recycle order life cycle.
package event

// Event recycle order event
type Event struct {
	Type  EvType
	Error error
}

// EvType recycle order event type
type EvType string

// definition of various event type
const (
	CommitFailed   EvType = "COMMIT_FAILED"
	CommitSuccess  EvType = "COMMIT_SUCCESS"
	DetectFailed   EvType = "DETECT_FAILED"
	DetectSuccess  EvType = "DETECT_SUCCESS"
	AuditApproved  EvType = "AUDIT_APPROVED"
	AuditRejected  EvType = "AUDIT_REJECTED"
	TransitFailed  EvType = "TRANSIT_FAILED"
	TransitSuccess EvType = "TRANSIT_SUCCESS"
	ReturnFailed   EvType = "RETURN_FAILED"
	ReturnHandling EvType = "RETURN_HANDLING"
	ReturnSuccess  EvType = "RETURN_SUCCESS"
)
