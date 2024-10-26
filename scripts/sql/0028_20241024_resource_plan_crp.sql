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
    SQLVER=0028,HCMVER=v1.6.9.2

    Notes:
    1. 修改资源预测单据表。
    2. 添加资源预测CRP关联表。
*/

START TRANSACTION;

# change unique key so that crp_demand_id can be repeated
ALTER TABLE `res_plan_crp_demand` DROP index `idx_uk_crp_demand_id`;
ALTER TABLE `res_plan_crp_demand` unique key `idx_uk_crp_demand_id_bk_biz_id` (`crp_demand_id`, `bk_biz_id`);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.6.9.2' as `hcm_ver`, '0028' as `sql_ver`;

COMMIT