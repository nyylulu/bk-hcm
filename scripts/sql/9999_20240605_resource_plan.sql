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
    `bk_biz_id`         bigint        not null comment '业务ID',
    `bsi_product_id`    bigint        not null comment '运营产品ID',
    `bsi_product_name`  varchar(64)   not null comment '运营产品名称',
    `plan_product_id`   bigint        not null comment '规划产品ID',
    `plan_product_name` varchar(64)   not null comment '规划产品名称',
    `virtual_dept_id`   bigint        not null comment '虚拟部门ID',
    `virtual_dept_name` varchar(64)   not null comment '虚拟部门名称',
    `demand_class`      varchar(64)   not null comment '预测的需求类型(枚举值：cvm(CVM)、ca(CA))',
    `os`                double        not null comment 'OS数，单位：台',
    `cpu_core`          double        not null comment '总CPU核心数，单位：核',
    `memory_size`       double        not null comment '总内存大小，单位：GB',
    `disk_size`         double        not null comment '总云盘大小，单位：GB',
    `itsm_sn`           varchar(64)            default '' comment '关联的itsm单据编码',
    `itsm_url`          varchar(64)            default '' comment '关联的itsm单据链接',
    `crp_sn`            varchar(64)            default '' comment '关联的crp单据编码',
    `crp_url`           varchar(64)            default '' comment '关联的crp单据链接',
    `remark`            varchar(1024) not null comment '预测说明，最少20字，最多1024字',
    `creator`           varchar(64)   not null comment '创建人',
    `reviser`           varchar(64)   not null comment '更新人',
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

insert into id_generator(`resource`, `max_id`)
values ('res_plan_demand', '0'),
       ('res_plan_demand_detail', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' AS `hcm_ver`, '9999' AS `sql_ver`;

COMMIT