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
 1. 支持OBS云账单 - zenlayer
 */
START TRANSACTION;



CREATE TABLE obs_zenlayer_bills (
    `id`                      varchar(64) not null,
    `vendor`                  varchar(64) not null,
    `main_account_id`         varchar(64) not null,
    `bill_year`               bigint not null,
    `bill_month`              tinyint not null,
    `set_index`               varchar(255) not null,

    yearMonth                 int,
    billing_main_account_id   varchar(256) ,
    billing_sub_account_id    varchar(384)  NOT NULL DEFAULT '',
    cost                      decimal(38, 10)  NOT NULL,
    currency                  varchar(64)  NOT NULL,
    rate                      double  NOT NULL,
    productid                 int  NOT NULL,
    `real_cost`               decimal(38, 10) not null,
    city                      varchar(64),
    contract_period           varchar(100) NOT NULL,
    description               varchar(255) NOT NULL,
    group_uid                 varchar(100) NOT NULL,
    pay_amount                decimal(50, 10),
    pay_content               varchar(100) NOT NULL,
    price                     decimal(50, 10),
    type                      varchar(64) NOT NULL,
    uid                       varchar(100) NOT NULL,
    zen_order_no              varchar(100) NOT NULL,
    accept_amount             decimal(50, 10),
    bill_monthly              varchar(64),
    cpu                       varchar(255) NOT NULL DEFAULT '',
    disk                      varchar(255) NOT NULL DEFAULT '',
    memory                    varchar(255) NOT NULL DEFAULT ''
);

insert into id_generator(`resource`, `max_id`)
values ('obs_zenlayer_bills', '0');


CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`,
       '9999'   as `sql_ver`;

COMMIT;