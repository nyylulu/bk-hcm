### 描述

- 该接口提供版本：v1.5.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询资源预测单据状态列表。

### URL

GET /api/v1/woa/plan/res_plan_ticket_status/list

### 输入参数

无

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "status": "init",
        "status_name": "待审批"
      },
      {
        "status": "auditing",
        "status_name": "审批中"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型         | 描述         |
|---------|--------------|------------|
| details | object array | 资源预测单据状态列表 |

#### data.details[n]

| 参数名称        | 参数类型   | 描述                                               |
|-------------|--------|--------------------------------------------------|
| status      | string | 单据状态（枚举值：init, auditing, rejected, done, failed） |
| status_name | string | 单据状态名称                                           |
