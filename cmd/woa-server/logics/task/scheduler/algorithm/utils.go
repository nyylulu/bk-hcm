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

// Package algorithm ...
package algorithm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"hcm/pkg/logs"
)

var (
	subversionRegExp = `(\d+)`
	subversionRe     = regexp.MustCompile(subversionRegExp)
)

// ValidateOSVersion validates os version
func ValidateOSVersion(requestVersion, hostVersion string) bool {
	if requestVersion == hostVersion {
		return true
	}

	requestVersionArray := strings.Split(requestVersion, "-")
	hostVersionArray := strings.Split(hostVersion, "-")
	if len(requestVersionArray) == 0 || len(requestVersionArray) != len(hostVersionArray) {
		return false
	}

	if requestVersionArray[0] != hostVersionArray[0] {
		return false
	}

	version1, err := parseVersion(requestVersionArray[len(requestVersionArray)-1], subversionRe, 1)
	if err != nil {
		logs.Errorf("parse request version (%s) failed: %v", requestVersion, err)
		return false
	}

	version2, err := parseVersion(hostVersionArray[len(hostVersionArray)-1], subversionRe, 1)
	if err != nil {
		logs.Errorf("parse host version (%s) failed: %v", hostVersion, err)
		return false
	}

	return version2[0] >= version1[0]
}

func parseVersion(version string, regex *regexp.Regexp, length int) ([]int, error) {
	matches := regex.FindAllStringSubmatch(version, -1)
	if len(matches) != 1 {
		return nil, fmt.Errorf("version string \"%v\" doesn't match expected regular expression: \"%v\"",
			version, regex.String())
	}
	versionArray := matches[0][1:]
	versions := make([]int, length)
	for index, versionStr := range versionArray {
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			return nil, fmt.Errorf("error while parsing \"%v\" in \"%v\"", versionStr, version)
		}
		versions[index] = version
	}
	return versions, nil
}
