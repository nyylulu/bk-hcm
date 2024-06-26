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

package enumor

// TCloudZiyan  腾讯自研云 厂商常量
const TCloudZiyan Vendor = "tcloud-ziyan"

func init() {
	RegisterVendor(TCloudZiyan, VendorInfo{
		NameEn:             "Tencent Cloud Ziyan Account",
		NameZh:             "腾讯云自研账号",
		MainAccountIDField: "cloud_main_account_id",
		SecretKeyField:     "cloud_secret_key",
	})

}
