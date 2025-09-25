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
    SQLVER=0051,HCMVER=v1.8.5.6

    Notes:
    1. woa_device_type 表新增数据列 technical_class、device_type_class
*/

START TRANSACTION;

ALTER TABLE `woa_device_type` ADD COLUMN `technical_class` VARCHAR(64) NOT NULL COMMENT '技术分类' AFTER `memory`;
ALTER TABLE `woa_device_type` ADD COLUMN `device_type_class` VARCHAR(64) NOT NULL COMMENT '机型分类' AFTER `device_type`;

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.5.6' as `hcm_ver`, '0051' as `sql_ver`;

COMMIT;
