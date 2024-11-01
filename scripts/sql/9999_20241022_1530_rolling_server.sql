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
    1. 添加滚服申请记录表。
    2. 添加滚服回收记录表。
    3. 添加滚服罚金明细表。
*/

START TRANSACTION;

# rolling_quota_config table structure
create table if not exists `rolling_quota_config`
(
    `id`          varchar(64) not null comment '唯一ID',
    `bk_biz_id`   bigint      not null comment '业务ID',
    `bk_biz_name` varchar(64) not null comment '业务名称',
    `year`        bigint      not null comment '配额应用年份',
    `month`       tinyint(1)  not null comment '配额应用月份',
    `quota`       bigint      not null comment 'cpu核心配额',
    `creator`     varchar(64) not null comment '创建人',
    `reviser`     varchar(64) not null comment '修改人',
    `created_at`  timestamp   not null default current_timestamp comment '该记录的创建时间',
    `updated_at`  timestamp   not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_year_month_bk_biz_id` (`year`, `month`, `bk_biz_id`)
) engine = innodb
  default charset = utf8mb4 comment '滚服基础额度配置表';

# rolling_quota_offset table structure
create table if not exists `rolling_quota_offset`
(
    `id`           varchar(64) not null comment '唯一ID',
    `bk_biz_id`    bigint      not null comment '业务ID',
    `bk_biz_name`  varchar(64) not null comment '业务名称',
    `year`         bigint      not null comment '配额应用年份',
    `month`        tinyint(1)  not null comment '配额应用月份',
    `quota_offset` bigint      not null comment 'cpu核心配额偏移量',
    `creator`      varchar(64) not null comment '创建人',
    `reviser`      varchar(64) not null comment '修改人',
    `created_at`   timestamp   not null default current_timestamp comment '该记录的创建时间',
    `updated_at`   timestamp   not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_year_month_bk_biz_id` (`year`, `month`, `bk_biz_id`),
    key `idx_reviser` (`reviser`)
) engine = innodb
  default charset = utf8mb4 comment '滚服配额偏移表';

# rolling_quota_offset_audit table structure
create table if not exists `rolling_quota_offset_audit`
(
    `id`               varchar(64) not null comment '唯一ID',
    `offset_config_id` varchar(64) not null comment '配额偏移表唯一ID',
    `operator`         varchar(64) not null comment '操作人',
    `quota_offset`     bigint      not null comment 'cpu核心配额偏移量',
    `rid`              varchar(64) not null comment '请求唯一标识',
    `app_code`         varchar(64) not null comment '应用的唯一标识',
    `created_at`       timestamp   not null default current_timestamp comment '该记录的创建时间',
    `updated_at`       timestamp   not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    key `idx_offset_config_id` (`offset_config_id`),
    key `idx_operator` (`operator`),
    key `idx_app_code` (`app_code`)
) engine = innodb
  default charset = utf8mb4 comment '滚服配额偏移审计表';

# rolling_global_config table structure
create table if not exists `rolling_global_config`
(
    `id`           varchar(64)     not null comment '唯一ID',
    `global_quota` bigint          not null comment 'CPU核心全局总配额',
    `biz_quota`    bigint          not null comment '单业务CPU核心基础配额',
    `unit_price`   decimal(38, 10) not null comment 'CPU核算单价（核/天）',
    `creator`      varchar(64)     not null comment '创建人',
    `reviser`      varchar(64)     not null comment '修改人',
    `created_at`   timestamp       not null default current_timestamp comment '该记录的创建时间',
    `updated_at`   timestamp       not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4 comment '滚服全局额度配置表';

# resource_pool_business table structure
create table if not exists `resource_pool_business`
(
    `id`          varchar(64) not null comment '唯一ID',
    `bk_biz_id`   bigint      not null comment '业务ID',
    `bk_biz_name` varchar(64) not null comment '业务名称',
    `creator`     varchar(64) not null comment '创建人',
    `reviser`     varchar(64) not null comment '修改人',
    `created_at`  timestamp   not null default current_timestamp comment '该记录的创建时间',
    `updated_at`  timestamp   not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_bk_biz_id` (`bk_biz_id`)
) engine = innodb
  default charset = utf8mb4 comment '资源池业务表';

# rolling_applied_record table structure
create table if not exists `rolling_applied_record`
(
    `id`             varchar(64)     not null comment '唯一ID',
    `applied_type`   varchar(16)     not null comment '申请类型(枚举值：normal-普通申请、resource_pool-资源池申请、cvm_product-管理员cvm生产)',
    `bk_biz_id`      bigint          not null comment '业务ID',
    `order_id`       bigint          not null comment '主机申请的订单号',
    `suborder_id`    varchar(64)     not null comment '主机申请的子订单号',
    `year`           bigint          not null comment '申请时间年份',
    `month`          tinyint(1)      not null comment '申请时间月份',
    `day`            tinyint(1)      not null comment '申请时间天',
    `roll_date`      int unsigned    not null comment '申请时间年月日',
    `applied_core`   bigint unsigned not null comment 'cpu申请核心数',
    `delivered_core` bigint unsigned not null comment 'cpu交付核心数',
    "instance_group" varchar(64)     not null comment '机型族',
    `creator`        varchar(64)     not null comment '创建人',
    `created_at`     timestamp       not null default current_timestamp comment '该记录的创建时间',
    `updated_at`     timestamp       not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_suborder_id_bk_biz_id_applied_type` (`suborder_id`, `bk_biz_id`, `applied_type`),
    KEY `idx_bk_biz_id_year_month_day` (`bk_biz_id`, `year`, `month`, `day`),
    KEY `idx_bk_biz_id_roll_date_applied_type` (`bk_biz_id`, `roll_date`, `applied_type`)
) engine = innodb
  default charset = utf8mb4 comment '滚服申请记录表';

# rolling_returned_record table structure
create table if not exists `rolling_returned_record`
(
    `id`                 varchar(64)     not null comment '唯一ID',
    `bk_biz_id`          bigint          not null comment '业务ID',
    `order_id`           bigint          not null comment '主机回收的订单号',
    `suborder_id`        varchar(64)     not null comment '主机回收的子订单号',
    `applied_record_id`  varchar(64)              default null comment '滚服申请执行情况表唯一标识',
    `match_applied_core` bigint unsigned not null comment '用于记录该回收子单，有多少核，用于归还匹配的申请子单',
    `year`               bigint          not null comment '申请时间年份',
    `month`              tinyint(1)      not null comment '申请时间月份',
    `day`                tinyint(1)      not null comment '申请时间天',
    `roll_date`          int unsigned    not null comment '申请时间年月日',
    `returned_way`       varchar(64)     not null comment '退还方式(枚举值：crp-通过crp退还、resource_pool-通过转移到资源池退还)',
    "instance_group"     varchar(64)     not null comment '机型族',
    `creator`            varchar(64)     not null comment '创建人',
    `created_at`         timestamp       not null default current_timestamp comment '该记录的创建时间',
    `updated_at`         timestamp       not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_suborder_id_bk_biz_id_returned_way` (`suborder_id`, `bk_biz_id`, `returned_way`),
    KEY `idx_bk_biz_id_applied_record_id` (`bk_biz_id`, `applied_record_id`),
    KEY `idx_bk_biz_id_year_month_day` (`bk_biz_id`, `year`, `month`, `day`),
    KEY `idx_bk_biz_id_roll_date` (`bk_biz_id`, `roll_date`)
) engine = innodb
  default charset = utf8mb4 comment '滚服回收记录表';

# 滚服罚金明细表
create table if not exists `rolling_fine_detail`
(
    `id`                varchar(64)     not null,
    `bk_biz_id`         bigint          not null comment '业务ID',
    `applied_record_id` varchar(64)     not null comment '滚服申请执行情况表唯一标识',
    `order_id`          bigint          not null comment '订单号',
    `suborder_id`       varchar(64)     not null comment '子订单号',
    `year`              bigint(1)       not null comment '子单号记录罚金的年份',
    `month`             tinyint(1)      not null comment '子单号记录罚金的月份',
    `day`               tinyint(1)      not null comment '子单号记录罚金的天',
    `roll_date`         int unsigned    not null comment '子单号记录罚金的年月日',
    `delivered_core`    bigint          not null comment 'cpu交付核心数',
    `returned_core`     bigint          not null comment 'cpu已退还核数',
    `fine`              decimal(38, 10) not null comment '超时退还罚金',
    `creator`           varchar(64)     not null,
    `created_at`        timestamp       not null default current_timestamp,
    primary key (`id`),
    unique key `idx_uk_year_month_day_applied_record_id` (`year`, `month`, `day`, `applied_record_id`),
    unique key `idx_uk_year_month_day_suborder_id` (`year`, `month`, `day`, `suborder_id`),
    key `idx_bk_biz_id_roll_date` (`bk_biz_id`, `roll_date`),
    key `idx_bk_biz_id_year_month_day` (`bk_biz_id`, `year`, `month`, `day`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='滚服罚金明细表';

insert into id_generator(`resource`, `max_id`)
values ('rolling_quota_config', '0'),
       ('rolling_quota_offset', '0'),
       ('rolling_quota_offset_audit', '0'),
       ('rolling_global_config', '0'),
       ('resource_pool_business', '0'),
       ('rolling_applied_record', '0'),
       ('rolling_returned_record', '0'),
       ('rolling_fine_detail', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT