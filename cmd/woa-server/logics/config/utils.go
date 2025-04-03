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

// Package config config
package config

import "fmt"

var regionToVpc = map[string]string{
	"ap-guangzhou":       "vpc-03nkx9tv",
	"ap-tianjin":         "vpc-1yoew5gc",
	"ap-shanghai":        "vpc-2x7lhtse",
	"eu-frankfurt":       "vpc-38klpz7z",
	"ap-singapore":       "vpc-706wf55j",
	"ap-tokyo":           "vpc-8iple1iq",
	"ap-seoul":           "vpc-99wg8fre",
	"ap-hongkong":        "vpc-b5okec48",
	"na-toronto":         "vpc-drefwt2v",
	"ap-xian-ec":         "vpc-efw4kf6r",
	"ap-nanjing":         "vpc-fb7sybzv",
	"ap-chongqing":       "vpc-gelpqsur",
	"ap-shenzhen":        "vpc-kwgem8tj",
	"na-siliconvalley":   "vpc-n040n5bl",
	"ap-hangzhou-ec":     "vpc-puhasca0",
	"ap-fuzhou-ec":       "vpc-hdxonj2q",
	"ap-wuhan-ec":        "vpc-867lsj6w",
	"ap-beijing":         "vpc-bhb0y6g8",
	"ap-jinan-ec":        "vpc-kgepmcdd",
	"ap-chengdu":         "vpc-r1wicnlq",
	"ap-zhengzhou-ec":    "vpc-54mjeaf8",
	"ap-shenyang-ec":     "vpc-rea7a2kc",
	"ap-changsha-ec":     "vpc-erdqk82h",
	"ap-hefei-ec":        "vpc-e0a5jxa7",
	"ap-shijiazhuang-ec": "vpc-6b3vbija",
}

// GetDftCvmVpc gets the default vpc of a region
func GetDftCvmVpc(region string) (string, error) {
	vpcID, ok := regionToVpc[region]
	if !ok {
		return "", fmt.Errorf("found no vpc with region: %s", region)
	}

	return vpcID, nil
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

// SecGroup network security group
type SecGroup struct {
	SecurityGroupId   string `json:"securityGroupId"`
	SecurityGroupName string `json:"securityGroupName"`
	SecurityGroupDesc string `json:"securityGroupDesc"`
}

var regionToSecGroup = map[string]*SecGroup{
	"ap-guangzhou": {
		SecurityGroupId:   "sg-ka67ywe9",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "腾讯自研上云-默认安全组",
	},
	"ap-tianjin": {
		SecurityGroupId:   "sg-c28492qp",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-shanghai": {
		SecurityGroupId:   "sg-ibqae0te",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "腾讯自研上云-默认安全组",
	},
	"eu-frankfurt": {
		SecurityGroupId:   "sg-cet13de0",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "云梯默认安全组",
	},
	"ap-singapore": {
		SecurityGroupId:   "sg-hjtqedoe",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-tokyo": {
		SecurityGroupId:   "sg-o1lfldnk",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "云梯默认安全组",
	},
	"ap-seoul": {
		SecurityGroupId:   "sg-i7h8hv5r",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "云梯默认安全组",
	},
	"ap-hongkong": {
		SecurityGroupId:   "sg-59kfufmn",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"na-toronto": {
		SecurityGroupId:   "sg-7l82d7km",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-xian-ec": {
		SecurityGroupId:   "sg-o4bmz4kg",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-nanjing": {
		SecurityGroupId:   "sg-dybs7i3y",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "腾讯自研上云-默认安全组",
	},
	"ap-chongqing": {
		SecurityGroupId:   "sg-l5usnzxw",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-shenzhen": {
		SecurityGroupId:   "sg-qkfewp0u",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"na-siliconvalley": {
		SecurityGroupId:   "sg-q7usygae",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-hangzhou-ec": {
		SecurityGroupId:   "sg-4ezyvbvl",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-fuzhou-ec": {
		SecurityGroupId:   "sg-leqa6w29",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-wuhan-ec": {
		SecurityGroupId:   "sg-p5ld4xyq",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-beijing": {
		SecurityGroupId:   "sg-rjwj7cnt",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-jinan-ec": {
		SecurityGroupId:   "sg-eag5dvzm",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-chengdu": {
		SecurityGroupId:   "sg-g504fnlx",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-zhengzhou-ec": {
		SecurityGroupId:   "sg-mdzp3pem",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-shenyang-ec": {
		SecurityGroupId:   "sg-jvdlgqyx",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-changsha-ec": {
		SecurityGroupId:   "sg-fohw41u4",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-hefei-ec": {
		SecurityGroupId:   "sg-qjn542yi",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
	"ap-shijiazhuang-ec": {
		SecurityGroupId:   "sg-5qwjawx2",
		SecurityGroupName: "云梯默认安全组",
		SecurityGroupDesc: "",
	},
}

// GetCvmDftSecGroup ...
func GetCvmDftSecGroup(region string) (*SecGroup, error) {
	sg, ok := regionToSecGroup[region]
	if !ok {
		return nil, fmt.Errorf("found no security group with region %s", region)
	}

	return sg, nil
}
