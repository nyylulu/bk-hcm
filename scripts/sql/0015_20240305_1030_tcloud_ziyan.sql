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
    SQLVER=0015,HCMVER=v1.4.0

    Notes:
    1. 增加自研云安全组规则表
    2. 添加申请单来源字段
*/

START TRANSACTION;


create table if not exists `tcloud_ziyan_security_group_rule`
(
    `id`                             varchar(64)  not null,
    `cloud_policy_index`             bigint(1)    not null,
    `type`                           varchar(20)  not null,
    `cloud_security_group_id`        varchar(255) not null,
    `security_group_id`              varchar(64)  not null,
    `account_id`                     varchar(64)  not null,
    `region`                         varchar(20)  not null,
    `version`                        varchar(255) not null,
    `action`                         varchar(10)  not null,
    `protocol`                       varchar(10)           default null,
    `port`                           varchar(255)          default null,
    `service_id`                     varchar(64)           default null,
    `cloud_service_id`               varchar(255)          default null,
    `service_group_id`               varchar(64)           default null,
    `cloud_service_group_id`         varchar(255)          default null,
    `ipv4_cidr`                      varchar(255)          default null,
    `ipv6_cidr`                      varchar(255)          default null,
    `cloud_target_security_group_id` varchar(255)          default null,
    `address_id`                     varchar(64)           default null,
    `cloud_address_id`               varchar(255)          default null,
    `address_group_id`               varchar(64)           default null,
    `cloud_address_group_id`         varchar(255)          default null,
    `memo`                           varchar(255)          default null,
    `creator`                        varchar(64)  not null,
    `reviser`                        varchar(64)  not null,
    `created_at`                     timestamp    not null default current_timestamp,
    `updated_at`                     timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_security_group_id_cloud_policy_index_type` (`cloud_security_group_id`, `cloud_policy_index`, `type`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='腾讯自研云安全组表';

create table if not exists `tcloud_ziyan_region`
(
    `id`          varchar(64) not null,
    `vendor`      varchar(16) not null,
    `region_id`   varchar(32) not null,
    `region_name` varchar(64) not null,
    `status`      varchar(32)          default '',
    `creator`     varchar(64)          default '',
    `reviser`     varchar(64)          default '',
    `created_at`  timestamp   not null default current_timestamp,
    `updated_at`  timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_region_id_status` (`region_id`, `status`),
    key `idx_uk_vendor` (`vendor`)
) engine = innodb
  default charset = utf8mb4   collate = utf8mb4_bin  comment ='腾讯自研云支持的地区列表';

insert into id_generator(`resource`, `max_id`)
values ('tcloud_ziyan_security_group_rule', '0'),
       ('tcloud_ziyan_region', '0');



CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.4.0' as `hcm_ver`, '0015' as `sql_ver`;

COMMIT