/*
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
    1. 修改`tcloud_lb_url_rule`表，增加`bk_biz_id`和`account_id`字段
*/

START TRANSACTION;

alter table `tcloud_lb_url_rule`
    add column `bk_biz_id` bigint not null default 0 COMMENT '业务ID' after `cloud_lb_id`,
add index `idx_bk_biz_id` (`bk_biz_id`);

alter table "tcloud_lb_url_rule"
    add column "account_id" varchar(64) not null default '' COMMENT '账号ID' after `bk_biz_id`;
add index "idx_account_id" ("account_id");

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;