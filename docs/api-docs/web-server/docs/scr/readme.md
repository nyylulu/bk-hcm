计算资源管理平台（SCR）在蓝鲸体系下，为业务提供计算资源的申请与回收生命周期管理能力。在提升资源运营效率的同时，规范资源管理流程，辅助业务提效降本。

## SCR迁移的URL规范说明

1、把原有的接口前缀【/api/v1/】统一替换为【/api/v1/woa/】

示例：
/api/v1/task/get/apply/ticket/audit  ====>  /api/v1/woa/task/get/apply/ticket/audit

2、如果原有的接口PATH里面包含了版本号，如：/api/v1/task/audit/v1/apply/ticket

示例：
/api/v1/task/audit/v1/apply/ticket  ====>  /api/v1/woa/task/audit/apply/ticket
