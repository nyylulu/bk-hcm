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
*/

START TRANSACTION;

# rolling_applied_record table structure
create table if not exists `rolling_applied_record`
(
    `id`                varchar(64) not null comment '唯一ID',
    `applied_type`      varchar(16) not null comment '申请类型(枚举值：normal-普通申请、resource_pool-资源池申请、cvm_product-管理员cvm生产)',
    `bk_biz_id`         bigint      not null comment '业务ID',
    `order_id`          varchar(64) not null comment '主机申请的订单号',
    `suborder_id`       varchar(64) not null comment '主机申请的子订单号',
    `year`              bigint      not null comment '申请时间年份',
    `month`             tinyint(1)  not null comment '申请时间月份',
    `day`               tinyint(1)  not null comment '申请时间天',
    `applied_core`      bigint      unsigned not null comment 'cpu申请核心数',
    `delivered_core`    bigint      unsigned not null comment 'cpu交付核心数',
    `creator`           varchar(64) not null comment '创建人',
    `created_at`        timestamp   not null default current_timestamp comment '该记录的创建时间',
    `updated_at`        timestamp   not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_suborder_id_bk_biz_id_applied_type` (`suborder_id`, `bk_biz_id`, `applied_type`),
    KEY `idx_bk_biz_id_year_month_day` (`bk_biz_id`, `year`, `month`, `day`)
) engine = innodb
  default charset = utf8mb4 comment '滚服申请记录表';

# rolling_returned_record table structure
create table if not exists `rolling_returned_record`
(
    `id`                    varchar(64) not null comment '唯一ID',
    `bk_biz_id`             bigint      not null comment '业务ID',
    `order_id`              varchar(64) not null comment '主机回收的订单号',
    `suborder_id`           varchar(64) not null comment '主机回收的子订单号',
    `applied_record_id`     varchar(64) default null comment '滚服申请执行情况表唯一标识',
    `match_applied_core`    bigint      unsigned not null comment '用于记录该回收子单，有多少核，用于归还匹配的申请子单',
    `year`                  bigint      not null comment '申请时间年份',
    `month`                 tinyint(1)  not null comment '申请时间月份',
    `day`                   tinyint(1)  not null comment '申请时间天',
    `returned_way`          varchar(64) not null comment '退还方式(枚举值：crp-通过crp退还、resource_pool-通过转移到资源池退还)',
    `creator`               varchar(64) not null comment '创建人',
    `created_at`            timestamp   not null default current_timestamp comment '该记录的创建时间',
    `updated_at`            timestamp   not null default CURRENT_TIMESTAMP on update CURRENT_TIMESTAMP comment '该记录的更新时间',
    primary key (`id`),
    unique key `idx_uk_suborder_id_bk_biz_id_returned_way` (`suborder_id`, `bk_biz_id`, `returned_way`),
    KEY `idx_bk_biz_id_applied_record_id` (`bk_biz_id`, `applied_record_id`),
    KEY `idx_bk_biz_id_year_month_day` (`bk_biz_id`, `year`, `month`, `day`)
) engine = innodb
  default charset = utf8mb4 comment '滚服回收记录表';

insert into id_generator(`resource`, `max_id`)
values ('rolling_applied_record', '0'),
       ('rolling_returned_record', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT