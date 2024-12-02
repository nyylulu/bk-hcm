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
    1. 添加预测需求表。
    2. 添加罚金计算基数表。
    3. 添加预测变更记录表。
*/

START TRANSACTION;

drop table if exists `res_plan_demand`;
# res_plan_demand table structure
create table if not exists `res_plan_demand`
(
    `id`                varchar(64)    not null comment '唯一ID',
    `locked`            tinyint        not null comment '是否已锁定(枚举值：0(未锁定)、1(已锁定))',
    `bk_biz_id`         bigint         not null comment '业务ID',
    `bk_biz_name`       varchar(64)    not null comment '业务名称',
    `op_product_id`     bigint         not null comment '运营产品ID',
    `op_product_name`   varchar(64)    not null comment '运营产品名称',
    `plan_product_id`   bigint         not null comment '规划产品ID',
    `plan_product_name` varchar(64)    not null comment '规划产品名称',
    `virtual_dept_id`   bigint         not null comment '虚拟部门ID',
    `virtual_dept_name` varchar(64)    not null comment '虚拟部门名称',
    `demand_class`      varchar(16)    not null comment '预测的需求类型(枚举值：CVM、CA)',
    `obs_project`       varchar(64)    not null comment '项目类型',
    `expect_time`       int unsigned   not null comment '期望到货时间，YYYYMMDD',
    `plan_type`         varchar(16)    not null comment '预测内外（枚举值：in_plan、out_plan）',
    `area_id`           varchar(64)    not null comment '地域ID',
    `area_name`         varchar(64)    not null comment '地域名称',
    `region_id`         varchar(64)    not null comment '地区/城市ID',
    `region_name`       varchar(64)    not null comment '地区/城市名称',
    `zone_id`           varchar(64)    not null comment '可用区ID',
    `zone_name`         varchar(64)    not null comment '可用区名称',
    `device_family`     varchar(64)    not null comment '机型族',
    `device_class`      varchar(64)    not null comment '机型类型',
    `device_type`       varchar(64)    not null comment '机型规格',
    `core_type`         varchar(64)    not null comment '核心类型',
    `disk_type`         varchar(64)    not null comment '云磁盘类型',
    `disk_type_name`    varchar(64)    not null comment '云磁盘类型中文名',
    `os`                decimal(16, 6) not null comment '预测实例数',
    `cpu_core`          bigint         not null comment '预测核心数',
    `memory`            bigint         not null comment '预测内存数',
    `disk_size`         bigint         not null comment '预测总磁盘大小',
    `disk_io`           int            not null comment '磁盘IO(MB/s)',
    `creator`           varchar(64)    not null comment '创建人',
    `reviser`           varchar(64)    not null comment '修改人',
    `created_at`        timestamp      not null default current_timestamp comment '该记录的创建时间',
    `updated_at`        timestamp      not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_bk_biz_id_dimensions` (`bk_biz_id`, `plan_type`, `demand_class`, `obs_project`, `expect_time`, `region_id`, `zone_id`, `device_type`, `disk_type`, `disk_io`),
    key `idx_obs_project` (`obs_project`),
    key `idx_device_class` (`device_class`),
    key `idx_device_type` (`device_type`),
    key `idx_expect_time` (`expect_time`),
    key `idx_region_id` (`region_id`)
) engine = innodb
  default charset = utf8mb4 comment '资源预测需求表';

# res_plan_demand_penalty_base table structure
create table if not exists `res_plan_demand_penalty_base`
(
    `id`                varchar(64) not null comment '唯一ID',
    `year`              int         not null comment '需求所属年份',
    `month`             tinyint     not null comment '需求所属月份',
    `week`              tinyint     not null comment '需求所属周（当月内，1-5）（跨月归属上个月）',
    `year_week`         tinyint     not null comment '需求所属周（当年内，1-52）（跨年归属上一年）',
    `source`            varchar(64) not null comment '数据来源（枚举值：local、crp）',
    `bk_biz_id`         bigint      not null comment '业务ID',
    `bk_biz_name`       varchar(64) not null comment '业务名称',
    `op_product_id`     bigint      not null comment '运营产品ID',
    `op_product_name`   varchar(64) not null comment '运营产品名称',
    `plan_product_id`   bigint      not null comment '规划产品ID',
    `plan_product_name` varchar(64) not null comment '规划产品名称',
    `virtual_dept_id`   bigint      not null comment '虚拟部门ID',
    `virtual_dept_name` varchar(64) not null comment '虚拟部门名称',
    `area_name`         varchar(64) not null comment '地域名称',
    `device_family`     varchar(64) not null comment '机型族',
    `cpu_core`          bigint      not null comment '预测需求核心数（如果是人工调减的数据，可能为负数）',
    `created_at`        timestamp   not null default current_timestamp comment '该记录的创建时间',
    `updated_at`        timestamp   not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_bk_biz_id_year_week_dimensions` (`bk_biz_id`, `year_week`, `source`, `area_name`, `device_family`),
    key `idx_year_month` (`year`, `month`)
) engine = innodb
  default charset = utf8mb4 comment '罚金计算基数表';

# res_plan_demand_changelog table structure
create table if not exists `res_plan_demand_changelog`
(
    `id`               varchar(64)    not null comment '唯一ID',
    `demand_id`        varchar(64)    not null comment 'HCM预测需求表ID',
    `ticket_id`        varchar(64)    not null comment 'HCM订单ID',
    `crp_order_id`     varchar(64)    not null comment 'CRP系统订单号',
    `suborder_id`      varchar(64)    not null comment '主机申领子订单号',
    `type`             varchar(16)    not null comment '变更类型(枚举值：append(追加)、adjust(修改)、delete(删除)、expend(消耗))',
    `expect_time`      varchar(16)    not null comment '期望到货时间，YYYY-MM-DD（变更后）',
    `obs_project`      varchar(64)    not null comment '项目类型（变更后）',
    `region_name`      varchar(64)    not null comment '地区/城市名称（变更后）',
    `zone_name`        varchar(64)    not null comment '可用区名称（变更后）',
    `device_type`      varchar(64)    not null comment '机型规则（变更后）',
    `os_change`        decimal(16, 6) not null comment '实例变更数',
    `cpu_core_change`  bigint         not null comment '核心变更数',
    `memory_change`    bigint         not null comment '内存变更数',
    `disk_size_change` bigint         not null comment '磁盘变更量',
    `remark`           varchar(1024)  not null comment '需求备注',
    `created_at`       timestamp      not null default current_timestamp comment '该记录的创建时间',
    `updated_at`       timestamp      not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    key `idx_demand_id` (`demand_id`)
) engine = innodb
  default charset = utf8mb4 comment '预测变更记录表';

insert into id_generator(`resource`, `max_id`)
values ('res_plan_demand_penalty_base', '0'),
       ('res_plan_demand_changelog', '0');

update id_generator set `max_id` = 0 where `resource` = 'res_plan_demand';

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT