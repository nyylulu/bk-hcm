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

package config

import "fmt"

var regionToVpc = map[string]string{
	"ap-guangzhou":     "vpc-03nkx9tv",
	"ap-tianjin":       "vpc-1yoew5gc",
	"ap-shanghai":      "vpc-2x7lhtse",
	"eu-frankfurt":     "vpc-38klpz7z",
	"ap-singapore":     "vpc-706wf55j",
	"ap-tokyo":         "vpc-8iple1iq",
	"ap-seoul":         "vpc-99wg8fre",
	"ap-hongkong":      "vpc-b5okec48",
	"na-toronto":       "vpc-drefwt2v",
	"ap-xian-ec":       "vpc-efw4kf6r",
	"ap-nanjing":       "vpc-fb7sybzv",
	"ap-chongqing":     "vpc-gelpqsur",
	"ap-shenzhen":      "vpc-kwgem8tj",
	"na-siliconvalley": "vpc-n040n5bl",
	"ap-hangzhou-ec":   "vpc-puhasca0",
	"ap-fuzhou-ec":     "vpc-hdxonj2q",
}

// GetDftCvmVpc gets the default vpc of a region
func GetDftCvmVpc(region string) (string, error) {
	vpc, ok := regionToVpc[region]
	if !ok {
		return "", fmt.Errorf("found no vpc with region %s", region)
	}

	return vpc, nil
}

// IsDftCvmVpc check if given vpc is the default vpc of a region
func IsDftCvmVpc(vpc string) bool {
	for _, val := range regionToVpc {
		if vpc == val {
			return true
		}
	}

	return false
}
