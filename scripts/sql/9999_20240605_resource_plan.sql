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
    1. 添加资源预测需求表。
    2. 添加资源预测需求详情表。
*/

START TRANSACTION;

# busniess resource plan ticket related table structure
create table if not exists `res_plan_ticket`
(
    `id`                varchar(64)   not null comment '唯一ID',
    `applicant`         varchar(64)   not null comment '申请人',
    `bk_biz_id`         bigint        not null comment '业务ID',
    `bk_product_id`     bigint        not null comment '运营产品ID',
    `bk_product_name`   varchar(64)   not null comment '运营产品名称',
    `plan_product_id`   bigint        not null comment '规划产品ID',
    `plan_product_name` varchar(64)   not null comment '规划产品名称',
    `virtual_dept_id`   bigint        not null comment '虚拟部门ID',
    `virtual_dept_name` varchar(64)   not null comment '虚拟部门名称',
    `demand_class`      varchar(16)   not null comment '预测的需求类型(枚举值：CVM、CA)',
    `os`                bigint        not null comment 'OS数，单位：台',
    `cpu_core`          bigint        not null comment '总CPU核心数，单位：核',
    `memory`            bigint        not null comment '总内存大小，单位：GB',
    `disk_size`         bigint        not null comment '总云盘大小，单位：GB',
    `remark`            varchar(1024) not null comment '预测说明，最少20字，最多1024字',
    `creator`           varchar(64)   not null comment '创建人',
    `reviser`           varchar(64)   not null comment '更新人',
    `submitted_at`      timestamp     not null default current_timestamp comment '提单或改单的时间',
    `created_at`        timestamp     not null default current_timestamp comment '该记录创建的时间',
    `updated_at`        timestamp     not null default current_timestamp on update current_timestamp comment '该记录更新的时间',
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4;

# busniess resource plan demand related table structure
create table if not exists `res_plan_demand`
(
    `id`            varchar(64)  not null comment '唯一ID',
    `ticket_id`     varchar(64)  not null comment '对应的单据唯一ID',
    `obs_project`   varchar(64)  not null comment 'OBS项目类型',
    `expect_time`   varchar(64)  not null comment '期望交付时间，格式为YYYY-MM-DD，例如2024-01-01',
    `area`          varchar(64)  not null comment '区域',
    `region`        varchar(64)  not null comment '城市',
    `zone`          varchar(64)  not null comment '可用区',
    `demand_source` varchar(64)  not null comment '需求分类/变更原因',
    `remark`        varchar(255) not null comment '需求备注',
    `cvm`           json                  default null comment '申请的CVM信息',
    `cbs`           json                  default null comment '申请的CBS信息',
    `creator`       varchar(64)  not null comment '创建人',
    `reviser`       varchar(64)  not null comment '更新人',
    `created_at`    timestamp    not null default current_timestamp comment '该记录创建的时间',
    `updated_at`    timestamp    not null default current_timestamp on update current_timestamp comment '该记录更新的时间',
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4;

# busniess resource plan ticket status info related table structure
create table if not exists `res_plan_ticket_status`
(
    `ticket_id`  varchar(64) not null comment '对应的单据唯一ID',
    `status`     varchar(64) not null comment '单据状态',
    `itsm_sn`    varchar(64)          default '' comment '关联的itsm单据编码',
    `itsm_url`   varchar(64)          default '' comment '关联的itsm单据链接',
    `crp_sn`     varchar(64)          default '' comment '关联的crp单据编码',
    `crp_url`    varchar(64)          default '' comment '关联的crp单据链接',
    `created_at` timestamp   not null default current_timestamp comment '该记录创建的时间',
    `updated_at` timestamp   not null default current_timestamp on update current_timestamp comment '该记录更新的时间',
    primary key (`ticket_id`)
) engine = innodb
  default charset = utf8mb4;

# woa zone info related table structure
create table if not exists `woa_zone`
(
    `id`          varchar(64) not null comment '唯一ID',
    `zone_id`     varchar(64) not null comment '可用区ID',
    `zone_name`   varchar(64) not null comment '可用区名称',
    `region_id`   varchar(64) not null comment '地区/城市ID',
    `region_name` varchar(64) not null comment '地区/城市名称',
    `area_id`     varchar(64) not null comment '地域ID',
    `area_name`   varchar(64) not null comment '地域名称',
    `created_at`  timestamp   not null default current_timestamp comment '该记录创建的时间',
    `updated_at`  timestamp   not null default current_timestamp on update current_timestamp comment '该记录更新的时间',
    primary key (`id`),
    unique key `idx_uk_zone` (`zone_id`)
) engine = innodb
  default charset = utf8mb4;

# woa device type info related table structure
create table if not exists `woa_device_type`
(
    `id`            varchar(64) not null comment '唯一ID',
    `device_type`   varchar(64) not null comment '机型',
    `device_class`  varchar(64) not null comment '机型分类',
    `device_family` varchar(64) not null comment '机型族',
    `core_type`     varchar(64) not null comment '核心类型',
    `cpu_core`      bigint(1)   not null comment 'CPU核心数',
    `memory`        bigint(1)   not null comment '内存大小，单位GB',
    `created_at`    timestamp   not null default current_timestamp comment '该记录创建的时间',
    `updated_at`    timestamp   not null default current_timestamp on update current_timestamp comment '该记录更新的时间',
    primary key (`id`),
    unique key `idx_uk_device_type` (`device_type`)
) engine = innodb
  default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('res_plan_ticket', '0'),
       ('res_plan_demand', '0'),
       ('woa_zone', '0'),
       ('woa_device_type', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' AS `hcm_ver`, '9999' AS `sql_ver`;

COMMIT