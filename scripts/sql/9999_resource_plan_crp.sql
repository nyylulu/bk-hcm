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
    1. 修改资源预测单据表。
    2. 添加资源预测CRP关联表。
*/

START TRANSACTION;

alter table res_plan_ticket
    rename column `bk_product_id` to `op_product_id`,
    rename column `bk_product_name` to `op_product_name`,
    rename column `os` to `updated_os`,
    rename column `cpu_core` to `updated_cpu_core`,
    rename column `memory` to `updated_memory`,
    rename column `disk_size` to `updated_disk_size`,
    add `type`               varchar(64) not null comment '单据类型(枚举值：add(新增)、adjust(修改)、cancel(取消))' after id,
    add `demands`            json        not null comment '需求列表，每个需求包括：original、updated两个部分' after type,
    add `original_os`        bigint      not null comment '原始OS数，单位：台',
    add `original_cpu_core`  bigint      not null comment '原始总CPU核心数，单位：核',
    add `original_memory`    bigint      not null comment '原始总内存大小，单位：GB',
    add `original_disk_size` bigint      not null comment '原始总云盘大小，单位：GB';

# resource plan crp demand related table structure
create table if not exists `res_plan_crp_demand`
(
    `id`                varchar(64) not null comment '唯一ID',
    `crp_demand_id`     bigint      not null comment 'CRP需求ID',
    `locked`            tinyint     not null comment '是否已锁定(枚举值：0(未锁定)、1(已锁定))',
    `demand_class`      varchar(16) not null comment '预测的需求类型(枚举值：CVM、CA)',
    `bk_biz_id`         bigint      not null comment '业务ID',
    `bk_biz_name`       varchar(64) not null comment '业务名称',
    `op_product_id`     bigint      not null comment '运营产品ID',
    `op_product_name`   varchar(64) not null comment '运营产品名称',
    `plan_product_id`   bigint      not null comment '规划产品ID',
    `plan_product_name` varchar(64) not null comment '规划产品名称',
    `virtual_dept_id`   bigint      not null comment '虚拟部门ID',
    `virtual_dept_name` varchar(64) not null comment '虚拟部门名称',
    `creator`           varchar(64) not null comment '创建人',
    `reviser`           varchar(64) not null comment '更新人',
    `created_at`        timestamp   not null default current_timestamp comment '该记录创建的时间',
    `updated_at`        timestamp   not null default current_timestamp on update current_timestamp comment '该记录更新的时间',
    primary key (`id`),
    unique key `idx_uk_crp_demand_id` (`crp_demand_id`)
) engine = innodb
  default charset = utf8mb4;

# resource plan penalty related table structure
create table if not exists `res_plan_penalty`
(
    `id`               varchar(64) not null comment '唯一ID',
    `op_product_id`    bigint      not null comment '运营产品ID',
    `year_month`       varchar(64) not null comment '罚金所属年月，格式为YYYY-MM',
    `penalty_cpu_core` DOUBLE      NOT NULL COMMENT '惩罚核心数',
    `creator`          varchar(64) not null comment '创建人',
    `reviser`          varchar(64) not null comment '更新人',
    `created_at`       timestamp   not null default current_timestamp comment '该记录创建的时间',
    `updated_at`       timestamp   not null default current_timestamp on update current_timestamp comment '该记录更新的时间',
    primary key (`id`),
    unique key `idx_uk_op_product_id_year_month` (`op_product_id`, `year_month`)
) engine = innodb
  default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('res_plan_crp_demand', '0'),
       ('res_plan_penalty', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT