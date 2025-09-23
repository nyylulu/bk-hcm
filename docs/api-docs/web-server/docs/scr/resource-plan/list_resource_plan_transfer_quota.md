### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：查询资源下资源预测转移额度列表。

### URL

GET /api/v1/woa/plans/resources/transfer_quotas/configs

### 输入参数

### 调用示例

```json
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "quota": 1000,
    "audit_quota": 5000
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型   | 描述                               |
|---------|-------- --|------------------------------------|
| code    | int       | 错误编码。 0表示success，>0表示失败错误 |
| message | string    | 请求失败返回的错误信息                 |
| data	  | object    | 响应数据                             |

#### data

| 参数名称     | 参数类型 | 描述               |
|-------------|--------|------------------|
| quota       | int    | 预测转移额度    |
| audit_quota | int    | 预测转移审批额度     |
