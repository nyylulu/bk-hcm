/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
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
 SQLVER=0002,HCMVER=v1.6.11.0
 
 Notes:
 1. obs滚服账单表
 */
START TRANSACTION;

create table if not exists `obs_rolling_bills`
(
    `id`                     varchar(64)  not null,
    `bk_biz_id`              bigint       not null comment '业务ID',
    `delivered_core`         bigint                DEFAULT 0 comment '已交付核心数',
    `returned_core`          bigint                DEFAULT 0 comment '已退还核心数',
    `not_returned_core`      bigint                DEFAULT 0 comment '未退还核心数',
    `year`                   bigint(1)    not null comment '记录账单的年份',
    `month`                  tinyint(1)   not null comment '记录账单的月份',
    `day`                    tinyint(1)   not null comment '记录账单的天',
    `roll_date`              int unsigned not null comment '记录账单的年月日',
    `creator`                varchar(64)  not null,
    `created_at`             timestamp    not null default current_timestamp,

    `data_date`              varchar(30)  NOT NULL comment '日期',
    `product_id`             int(11)               DEFAULT 0 comment '运营产品ID',
    `business_set_id`        int(11)               DEFAULT 0,
    `business_set_name`      varchar(30)           DEFAULT '' comment '业务集名称',
    `business_id`            int(11)               DEFAULT 0,
    `business_name`          varchar(30)           DEFAULT '' comment '业务名称',
    `business_mod_id`        int(11)               DEFAULT 0,
    `business_mod_name`      varchar(30)           DEFAULT '' comment '业务模块名称',
    `uin`                    varchar(30)           DEFAULT '',
    `app_id`                 varchar(30)           DEFAULT '',
    `user`                   varchar(30)           DEFAULT '' comment '使用人',
    `city_id`                int(11)               DEFAULT 10000 comment '城市ID',
    `campus_id`              int(11)               DEFAULT 0 comment '园区ID',
    `idc_unit_id`            int(11)               DEFAULT 0 comment '管理单元ID',
    `idc_unit_name`          varchar(30)           DEFAULT '' comment '管理单元名称',
    `module_id`              int(11)               DEFAULT 0,
    `module_name`            varchar(30)           DEFAULT '',
    `zone_id`                int(11)               DEFAULT 0 comment '可用区ID',
    `zone_name`              varchar(30)           DEFAULT '' comment '可用区名称',
    `platform_id`            int(11)               DEFAULT 0 comment '平台ID',
    `res_class_id`           int(11)               DEFAULT 0 comment '资源规格ID',
    `cluster_id`             varchar(30)           DEFAULT 0 comment '集群ID',
    `platform_res_id`        varchar(200)          DEFAULT '' comment '最小粒度资源ID',
    `bandwidth_type_id`      int(11)               DEFAULT 0 comment '带宽类型ID',
    `operator_name_id`       int(11)               DEFAULT 0 comment '运营商ID',
    `amount`                 double                DEFAULT 0 comment '核算用量',
    `amount_in_current_date` double                DEFAULT 0 comment '参考日用量',
    `cost`                   double                DEFAULT 0 comment '成本',
    `extend_detail`          varchar(30)           DEFAULT '' comment '扩展详情',
    primary key (`id`),
    unique key `idx_uk_bk_biz_id_year_month_day` (`bk_biz_id`, `year`, `month`, `day`),
    key `idx_bk_biz_id_roll_date` (`bk_biz_id`, `roll_date`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='obs滚服账单';

insert into id_generator(`resource`, `max_id`)
values ('obs_rolling_bills', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.6.11.0' as `hcm_ver`, '0002' as `sql_ver`;

COMMIT;
