### 描述

- 该接口提供版本：v1.7.0.0+。
- 该接口所需权限：无。
- 该接口功能描述：查询小额绿通的配置。

### URL

GET /api/v1/woa/green_channels/configs

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
    "biz_quota": 500,
    "ieg_quota": 50000,
    "audit_threshold": 1000
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

| 参数名称            | 参数类型 | 描述               |
|-----------------|------|------------------|
| biz_quota       | int  | 业务申请小额绿通的每周额度    |
| ieg_quota       | int  | 整个ieg申请小额绿通的每周额度 |
| audit_threshold | int  | 小额绿通自动审批核数上限     |
