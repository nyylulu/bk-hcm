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
    1. 新增`res_plan_transfer_applied_record`表
*/

START TRANSACTION;

create table if not exists `res_plan_transfer_applied_record` (
    `id` varchar(64) not null comment '唯一ID',
    `applied_type` varchar(64) not null comment '转移类型',
    `bk_biz_id` bigint not null comment '业务ID',
    `sub_ticket_id` varchar(64) not null comment '预测调整子单号',
    `year` bigint not null comment '预测调整时间-年',
    `technical_class` varchar(64) not null comment '技术分类',
    `obs_project` varchar(64) not null comment '项目类型',
    `expected_core` bigint not null comment '预期转移的核心数',
    `applied_core` bigint not null comment '成功转移的核心数',
    `creator` varchar(64) not null comment '创建者',
    `reviser` varchar(64) not null comment '更新者',
    `created_at` datetime not null DEFAULT CURRENT_TIMESTAMP comment '创建时间',
    `updated_at` datetime not null DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP comment '更新时间',
    primary key (`id`),
    key `idx_bk_biz_id` (`bk_biz_id`),
    unique key `idx_sub_ticket_id_technical_class_obs_project` (`sub_ticket_id`, `technical_class`, `obs_project`),
    key `idx_applied_type_year_technical_class_obs_project` (`applied_type`, `year`, `technical_class`, `obs_project`)
) engine=innodb default charset=utf8mb4 comment='预测转移额度执行记录表';


insert into id_generator(`resource`, `max_id`)
values ('res_plan_transfer_applied_record', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
