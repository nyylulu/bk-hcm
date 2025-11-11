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
    SQLVER=9999,HCMVER=v9.9.9.9

    Notes:
    1. 新增 CVM 申请单统计配置表 cvm_apply_order_statistics_config
*/

START TRANSACTION;

-- 1. CVM 申请单统计配置表
CREATE TABLE IF NOT EXISTS `cvm_apply_order_statistics_config`
(
    `id`           VARCHAR(64)  NOT NULL COMMENT '主键',
    `year_month`   VARCHAR(16)      NOT NULL COMMENT '配置所属月份，格式：YYYY-MM',
    `bk_biz_id`    BIGINT       NOT NULL COMMENT '业务ID',
    `sub_order_ids` TEXT NOT NULL DEFAULT '' COMMENT '子单号列表，最多100个，逗号分隔',
    `start_at`     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '开始时间，格式：YYYY-MM-DD',
    `end_at`       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '结束时间，格式：YYYY-MM-DD',
    `memo`         VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
    `extension`    JSON                  DEFAULT NULL COMMENT '扩展信息 JSON',
    `creator`      VARCHAR(64)  NOT NULL COMMENT '创建者',
    `reviser`      VARCHAR(64)  NOT NULL COMMENT '更新者',
    `created_at`   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_year_month` (`year_month`),
    KEY `idx_bk_biz_id` (`bk_biz_id`)
    ) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4
    COLLATE = utf8mb4_bin COMMENT ='CVM申请单统计配置表';

INSERT INTO `id_generator`(`resource`, `max_id`)
VALUES ('cvm_apply_order_statistics_config', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9.9' AS `hcm_ver`, '9999' AS `sql_ver`;

COMMIT;

