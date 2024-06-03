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
    1. 支持腾讯云负载均衡
*/

START TRANSACTION;


-- 1. 负载均衡七层规则
create table `tcloud_ziyan_lb_url_rule`
(
    `id`                    varchar(64)  not null,
    `cloud_id`              varchar(255) not null,
    `name`                  varchar(255) not null,
    `rule_type`             varchar(64)  not null,

    `lb_id`                 varchar(255) not null,
    `cloud_lb_id`           varchar(255) not null,
    `lbl_id`                varchar(255) not null,
    `cloud_lbl_id`          varchar(255) not null,
    `target_group_id`       varchar(255)          default '',
    `cloud_target_group_id` varchar(255)          default '',

    `domain`                varchar(255)          default '',
    `url`                   varchar(255)          default '',
    `scheduler`             varchar(64)  not null,
    `session_type`          varchar(64)           default '',
    `session_expire`        bigint                default 0,
    `health_check`          json                  default null,
    `certificate`           json                  default null,
    `memo`                  varchar(255)          default '',


    `creator`               varchar(64)  not null,
    `reviser`               varchar(64)  not null,
    `created_at`            timestamp    not null default current_timestamp,
    `updated_at`            timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id_lbl_id` (`cloud_id`, `lbl_id`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='负载均衡七层规则';


insert into id_generator(`resource`, `max_id`)
values ('tcloud_ziyan_lb_url_rule', '0');


CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT