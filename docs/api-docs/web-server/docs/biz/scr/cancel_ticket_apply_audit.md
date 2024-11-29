### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：取消主机申请的当前审批单据（ITSM审批单据或CRP审批单据）。

### URL

PATCH /api/v1/woa/task/apply/ticket/audit/cancel

### 输入参数

| 参数名称         | 参数类型   | 必选 | 描述   |
|--------------|--------|----|------|
| sub_order_id | string | 是  | 单据ID |

### 调用示例

```json
{
  "sub_order_id": "xxx"
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
