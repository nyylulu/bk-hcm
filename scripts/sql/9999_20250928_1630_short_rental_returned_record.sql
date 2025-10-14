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
    1. 新增short_rental_returned_record表
*/

START TRANSACTION;

CREATE TABLE IF NOT EXISTS `short_rental_returned_record` (
    `id` VARCHAR(64) NOT NULL COMMENT '唯一标识',
    `bk_biz_id` BIGINT NOT NULL COMMENT '业务ID',
    `bk_biz_name` VARCHAR(64) COMMENT '业务名称',
    `op_product_id` BIGINT COMMENT '运营产品ID',
    `op_product_name` VARCHAR(64) COMMENT '运营产品名称',
    `plan_product_id` BIGINT COMMENT '规划产品ID',
    `plan_product_name` VARCHAR(64) COMMENT '规划产品名称',
    `virtual_dept_id` BIGINT COMMENT '虚拟部门ID',
    `virtual_dept_name` VARCHAR(64) COMMENT '虚拟部门名称',
    `order_id` BIGINT COMMENT '主机回收的订单号',
    `suborder_id` VARCHAR(64) NOT NULL COMMENT '主机回收的子订单号',
    `year` BIGINT COMMENT '退回时间-年',
    `month` TINYINT COMMENT '退回时间-月',
    `returned_date` INT UNSIGNED COMMENT '退回时间-年月日（格式为YYYYMMDD）',
    `physical_device_family` VARCHAR(64) NOT NULL COMMENT '物理机机型族',
    `region_id` VARCHAR(64) NOT NULL COMMENT '地区/城市ID',
    `region_name` VARCHAR(64) COMMENT '地区/城市名称',
    `status` VARCHAR(64) NOT NULL COMMENT '回收状态（RETURNING/DONE/TERMINATE）',
    `returned_core` BIGINT COMMENT '回收核心数',
    `creator` VARCHAR(64) NOT NULL comment '创建人',
    `reviser` VARCHAR(64) NOT NULL comment '更新者',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_biz_suborder_device_family_region` (`bk_biz_id`, `suborder_id`, `physical_device_family`, `region_id`),
    KEY `idx_op_product_year_month_device_family_region` (`op_product_id`, `year`, `month`, `physical_device_family`,
                                                        `region_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='短租退回记录表';

insert into id_generator(`resource`, `max_id`)
values ('short_rental_returned_record', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
