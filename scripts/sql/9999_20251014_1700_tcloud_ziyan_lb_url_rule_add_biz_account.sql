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
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
    1. 为自研云负载均衡URL规则表添加业务ID和账号ID字段
*/

START TRANSACTION;

-- 为自研云负载均衡URL规则表添加业务ID和账号ID字段
ALTER TABLE `tcloud_ziyan_lb_url_rule` 
    ADD COLUMN `bk_biz_id` BIGINT NOT NULL DEFAULT 0 COMMENT '业务ID' AFTER `memo`,
    ADD COLUMN `account_id` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '账号ID' AFTER `bk_biz_id`,
    ADD INDEX `idx_bk_biz_id` (`bk_biz_id`),
    ADD INDEX `idx_account_id` (`account_id`);

-- 刷新历史数据的bk_biz_id和account_id
UPDATE `tcloud_ziyan_lb_url_rule` tz
    INNER JOIN `load_balancer` lb ON tz.lb_id = lb.id
    SET
        tz.bk_biz_id = lb.bk_biz_id,
        tz.account_id = lb.account_id
WHERE
    tz.bk_biz_id = 0 OR tz.account_id = '';

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
