### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询单据类型列表。

### URL

POST /api/v1/woa/metas/ticket_types/list

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
        "ticket_type": "add",
        "ticket_type_name": "新增"
      },
      {
        "ticket_type": "update",
        "ticket_type_name": "更新"
      },
      {
        "ticket_type": "cancel",
        "ticket_type_name": "取消"
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

| 参数名称    | 参数类型         | 描述     |
|---------|--------------|--------|
| details | object array | 单据类型列表 |

#### data.details[n]

| 参数名称             | 参数类型   | 描述                              |
|------------------|--------|---------------------------------|
| ticket_type      | string | 单据类型唯一标识(枚举值：add/update/cancel) |
| ticket_type_name | string | 单据类型名称                          |
