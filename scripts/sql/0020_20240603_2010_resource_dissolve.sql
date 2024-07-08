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
    SQLVER=0020,HCMVER=v1.6.0

    Notes:
    1. 添加裁撤服务的裁撤模块表
    2. 添加裁撤服务的裁撤主机表
*/

START TRANSACTION;

--  1. 裁撤模块表
create table if not exists `recycle_module_info`
(
    `id`                  varchar(64)  not null,
    `name`                varchar(255) not null,
    `start_time`          varchar(10)  not null,
    `end_time`            varchar(10)  not null,
    `which_stages`        tinyint      not null,
    `recycle_type`        tinyint      not null,
    `creator`             varchar(64)  not null,
    `reviser`             varchar(64)  not null,
    `created_at`          timestamp    not null default current_timestamp,
    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_name` (`name`)
    ) engine = innodb
    default charset = utf8mb4
    collate utf8mb4_bin comment ='裁撤模块表';

--  2. 裁撤主机表
create table if not exists `recycle_host_info`
(
    `id`                  varchar(64)  not null,
    `asset_id`            varchar(255) not null,
    `inner_ip`            varchar(15)  not null,
    `module`              varchar(255) not null,
    `creator`             varchar(64)  not null,
    `reviser`             varchar(64)  not null,
    `created_at`          timestamp    not null default current_timestamp,
    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_asset_id` (`asset_id`),
    unique key `idx_uk_inner_ip` (`inner_ip`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin comment ='裁撤主机表';

insert into id_generator(`resource`, `max_id`)
values ('recycle_module_info', '0'),
       ('recycle_host_info', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.6.0' as `hcm_ver`, '0020' as `sql_ver`;

COMMIT
