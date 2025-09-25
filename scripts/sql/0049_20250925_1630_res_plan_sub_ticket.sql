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
    1. 新增预测子单据表
    2. 预测单据表中cpu、内存、磁盘数应为int而不是float
*/

START TRANSACTION;

--  1. 预测子单据表
create table if not exists `res_plan_sub_ticket`
(
    `id`           varchar(64)  not null comment '主键',
    `ticket_id`    varchar(64)  not null comment '父单据ID',
    `sub_type`     varchar(64) not null comment '子单据类型',
    `sub_demands` json NOT NULL COMMENT '子单据需求列表，每个需求包括：original、updated两个部分',
    `bk_biz_id` bigint NOT NULL COMMENT '业务ID',
    `bk_biz_name` varchar(64) NOT NULL COMMENT '业务名称',
    `op_product_id` bigint NOT NULL COMMENT '运营产品ID',
    `op_product_name` varchar(64) NOT NULL COMMENT '运营产品名称',
    `plan_product_id` bigint NOT NULL COMMENT '规划产品ID',
    `plan_product_name` varchar(64) NOT NULL COMMENT '规划产品名称',
    `virtual_dept_id` bigint NOT NULL COMMENT '虚拟部门ID',
    `virtual_dept_name` varchar(64) NOT NULL COMMENT '虚拟部门名称',
    `status` varchar(64) NOT NULL COMMENT '子单据状态',
    `message` varchar(255) NOT NULL DEFAULT '' COMMENT '子单据信息',
    `stage` varchar(64) NOT NULL COMMENT '子单据审批阶段',
    `admin_audit_status` varchar(64) NOT NULL COMMENT '管理员审批结果',
    `admin_audit_operator` varchar(64) NOT NULL COMMENT '管理员审批人',
    `admin_audit_at` datetime NOT NULL COMMENT '管理员审批时间',
    `crp_sn` varchar(64) NOT NULL DEFAULT '' COMMENT 'CRP单据ID',
    `crp_url` varchar(64) NOT NULL DEFAULT '' COMMENT 'CRP单据审批链接',
    `sub_original_os` double NOT NULL COMMENT '原始OS数，单位：台',
    `sub_original_cpu_core` bigint NOT NULL COMMENT '原始总CPU核心数，单位：核',
    `sub_original_memory` bigint NOT NULL COMMENT '原始总内存大小，单位：GB',
    `sub_original_disk_size` bigint NOT NULL COMMENT '原始总云盘大小，单位：GB',
    `sub_updated_os` double NOT NULL COMMENT '更新后OS数，单位：台',
    `sub_updated_cpu_core` bigint NOT NULL COMMENT '更新后总CPU核心数，单位：核',
    `sub_updated_memory` bigint NOT NULL COMMENT '更新后总内存大小，单位：GB',
    `sub_updated_disk_size` bigint NOT NULL COMMENT '更新总云盘大小，单位：GB',
    `submitted_at` datetime NOT NULL COMMENT '提单或改单的时间',
    `creator` varchar(64) NOT NULL COMMENT '创建人',
    `reviser` varchar(64) NOT NULL COMMENT '更新人',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '该记录创建的时间',
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '该记录更新的时间',
    primary key (`id`),
    index `idx_ticket_id_bk_biz_id` (`ticket_id`, `bk_biz_id`)
    ) engine = innodb
    default charset = utf8mb4
    collate utf8mb4_bin comment ='预测子单据表';

-- 2. 预测单据表中cpu、内存、磁盘数应为int而不是float
alter table res_plan_ticket modify column updated_cpu_core bigint DEFAULT NULL;
alter table res_plan_ticket modify column updated_memory bigint DEFAULT NULL;
alter table res_plan_ticket modify column updated_disk_size bigint DEFAULT NULL;
alter table res_plan_ticket modify column original_cpu_core bigint DEFAULT NULL;
alter table res_plan_ticket modify column original_memory bigint DEFAULT NULL;
alter table res_plan_ticket modify column original_disk_size bigint DEFAULT NULL;

insert into id_generator(`resource`, `max_id`)
values ('res_plan_sub_ticket', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
