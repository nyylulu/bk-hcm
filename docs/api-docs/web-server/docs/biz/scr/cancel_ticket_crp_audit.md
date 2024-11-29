### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：取消主机申请子订单的所有可取消CRP单据。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/apply/ticket/crp_audit/cancel

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述       |
|-------------|--------|----|----------|
| suborder_id | string | 是  | HCM子订单ID |

### 调用示例

```json
{
  "suborder_id": "xxx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": ""
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
