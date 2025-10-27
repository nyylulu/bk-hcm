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
    SQLVER=0052,HCMVER=v1.8.7.0
    Notes:
    1. 为自研云CVM表添加bk_asset_id字段，从extension字段中提取固资号
    2. 清理extension字段中的bk_asset_id数据
*/

START TRANSACTION;

-- 添加bk_asset_id字段到cvm表
alter table cvm
    add column bk_asset_id varchar(64) DEFAULT '' COMMENT '固资号';

-- 添加索引
alter table cvm
    ADD INDEX idx_bk_asset_id (`bk_asset_id`, `id`);

-- 从extension字段中提取固资号到bk_asset_id字段
update cvm set bk_asset_id = JSON_UNQUOTE(JSON_EXTRACT(extension, '$.bk_asset_id')) 
where vendor = 'tcloud-ziyan'
  and JSON_EXTRACT(extension, '$.bk_asset_id') is not null 
  and JSON_EXTRACT(extension, '$.bk_asset_id') != 'null'
  and JSON_EXTRACT(extension, '$.bk_asset_id') != '';

-- 清理extension字段中的bk_asset_id数据
update cvm set extension = JSON_REMOVE(extension, '$.bk_asset_id')
where vendor = 'tcloud-ziyan'
  and JSON_EXTRACT(extension, '$.bk_asset_id') is not null
  and JSON_EXTRACT(extension, '$.bk_asset_id') != 'null'
  and JSON_EXTRACT(extension, '$.bk_asset_id') != '';

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.7.0' as `hcm_ver`, '0052' as `sql_ver`;

COMMIT;