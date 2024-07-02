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
 SQLVER=9999,HCMVER=v9.9.9
 
 Notes:
 1. 支持OBS云账单
 */
START TRANSACTION;

create table if not exists `id_generator`
(
    `resource` varchar(64) not null,
    `max_id`   varchar(64) not null,
    primary key (`resource`)
) engine = innodb
  default charset = utf8mb4;

create table `obs_huawei_bills` (
    `id` varchar(64) not null,
    `vendor` varchar(64) not null,
    `main_account_id` varchar(64) not null,
    `bill_year` bigint(1) not null,
    `bill_month` tinyint(1) not null,
    `set_index` varchar(255) not null,
    `effective_time` longtext not null,
    `expire_time` longtext not null,
    `product_id` longtext not null,
    `product_name` longtext not null,
    `order_id` longtext not null,
    `amount` longtext not null,
    `measure_id` longtext not null,
    `usage_type` longtext not null,
    `usages` longtext not null,
    `usage_measure_id` longtext not null,
    `free_resource_usage` longtext not null,
    `free_resource_measure_id` longtext not null,
    `cloud_service_type` longtext not null,
    `region` longtext not null,
    `resource_type` longtext not null,
    `charge_mode` longtext not null,
    `resource_tag` longtext not null,
    `resource_name` longtext not null,
    `resource_id` longtext not null,
    `bill_type` longtext not null,
    `enterprise_project_id` longtext not null,
    `period_type` longtext not null,
    `spot` longtext not null,
    `ri_usage` longtext not null,
    `ri_usage_measure_id` longtext not null,
    `official_amount` longtext not null,
    `discount_amount` longtext not null,
    `cash_amount` longtext not null,
    `credit_amount` longtext not null,
    `coupon_amount` longtext not null,
    `flexipurchase_coupon_amount` longtext not null,
    `stored_card_amount` longtext not null,
    `bonus_amount` longtext not null,
    `debt_amount` longtext not null,
    `adjustment_amount` longtext not null,
    `spec_size` longtext not null,
    `spec_size_measure_id` longtext not null,
    `account_name` longtext not null,
    `productid` int(11) not null,
    `account_type` longtext not null,
    `yearMonth` int(11) not null,
    `fetchTime` datetime not null,
    `total_count` int(11) not null,
    `rate` double not null,
    `real_cost` decimal(38, 10) not null,
    index `idx_bill_item` (`set_index`),
    index `idx_bill_item_delete` (
        `vendor`,
        `bill_year`,
        `bill_month`,
        `main_account_id`
    )
);

insert into id_generator(`resource`, `max_id`)
values ('obs_huawei_bills', '0');

CREATE
OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT
    'v9.9.9' as `hcm_ver`,
    '9999' as `sql_ver`;

COMMIT;