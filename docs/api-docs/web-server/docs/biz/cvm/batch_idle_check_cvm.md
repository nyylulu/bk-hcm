### 描述

- 该接口提供版本：v1.8.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：批量空闲检查。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/batch/idle_check

### 输入参数

| 参数名称       | 参数类型       | 必选 | 描述                         |
|---------------|--------------|------|-----------------------------|
| bk_biz_id     | int64        | 是   | 业务ID                       |
| bk_host_ids   | int64 array  | 是   | hostID列表,最多支持500个hostID |
| exclude_steps | string array | 否   | 需要排除的检查步骤(PRE_CHECK:检查CC模块和负责人、CHECK_UWORK:检查是否有Uwork故障或流程单据、CHECK_TCAPLUS:检查是否有Tcaplus记录、BASIC_CHECK:tmp,tgw,tgw nat,l5策略检查、CVM_CHECK:检查cvm, docker on cvm的安全组与CLB策略、CHECK_SAFETY:安全基线检查、CHECK_RETURN:检查是否有退回单据、CHECK_PROCESS:空闲检查、CHECK_OWNER:检查是否包含虚拟子机、CHECK_PM_OUTERIP:物理机回收外网IP、CHECK_DBM:检查DBM使用记录)             |

### 调用示例

```json
{
  "bk_host_ids": [
    435,
    265
  ],
  "exclude_steps": ["PRE_CHECK"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "task_management_id": "xxxxxx",
    "idle_check_suborder_id": "xxxxxx"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |


#### data参数说明

| 参数名称                 | 参数类型   | 描述        |
|----------------------|--------|-----------|
| task_management_id   | string | 任务管理id    |
| idle_check_suborder_id | string | 空闲检查时生成的子单号 |