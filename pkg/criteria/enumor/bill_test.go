/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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

package enumor

import (
	"testing"
)

// TestIsAIBillItem ...
func TestIsAIBillItem(t *testing.T) {
	// 定义测试用例表
	tests := []struct {
		name     string // 测试用例名称
		input    string // 输入字符串
		expected bool   // 预期结果
	}{
		{"Contains Gemini AI", "This is about Gemini AI", true},
		{"Contains claude", "claude is from anthropic", true},
		{"Invalid aagemini", "aagemini is not valid", false},
		{"Invalid geminiaa", "geminiaa is not valid", false},
		{"Valid with underscore prefix", "aa_gemini is valid", true},
		{"Valid with underscore suffix", "gemini_aa is valid", true},
		{"Invalid combined word", "GeminiClaude is not valid", false},
		{"Uppercase GEMINI", "GEMINI is valid", true},
		{"Uppercase CLAUDE", "CLAUDE is valid", true},
		{"No target words", "No target words here", false},
		{"Multiple target words", "Check gemini and claude", true},
		{"With hyphen", "gemini-claude", true},
		{"With slash", "gemini/claude", true},
		{"With number suffix", "gemini6", true},
		{"With number prefix", "6gemini", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 执行被测函数
			got := IsAIBillItem(tt.input)
			// 验证结果
			if got != tt.expected {
				t.Errorf("ContainsTargetWords(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
