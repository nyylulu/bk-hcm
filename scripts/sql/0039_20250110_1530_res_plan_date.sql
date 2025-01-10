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
    SQLVER=0039,HCMVER=v1.7.2.0

    Notes:
    1. 添加预测所属周校验表 res_plan_week
*/

START TRANSACTION;

--  1. 预测所属周校验表
create table if not exists `res_plan_week`
(
    `id`           varchar(64)  not null comment '主键',
    `year`         int          not null comment '需求所属年份',
    `month`        tinyint      not null comment '需求所属月份',
    `year_week`    tinyint      not null comment '需求所属周(当年内，1-52)',
    `start`        int unsigned not null comment '期望到货时间范围，YYYYMMDD',
    `end`          int unsigned not null comment '期望到货时间范围，YYYYMMDD',
    `is_holiday`   tinyint      not null comment '是否为节假日(枚举值：0(false)、1(true))',
    `created_at`   timestamp    not null default current_timestamp comment '创建时间',
    `updated_at`   timestamp    not null default current_timestamp on update current_timestamp comment '更新时间',
    primary key (`id`),
    unique key `idx_year_month_week` (`year`, `month`, `year_week`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin comment ='预测所属周校验表';

insert into id_generator(`resource`, `max_id`)
values ('res_plan_week', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.7.2.0' as `hcm_ver`, '0039' as `sql_ver`;

COMMIT;
