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

/*
    SQLVER=0041,HCMVER=v1.7.6.1

    Notes:
    1. 负载均衡七层规则表对目标组id添加索引
    2. 负载均衡七层规则表云上负载均衡id、云上监听器id添加联合索引
*/

START TRANSACTION;

create index idx_target_group_id ON tcloud_ziyan_lb_url_rule (`target_group_id`);
create index idx_cloud_lb_id_cloud_lbl_id ON tcloud_ziyan_lb_url_rule (`cloud_lb_id`, `cloud_lbl_id`);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.7.6.1' as `hcm_ver`, '0041' as `sql_ver`;

COMMIT