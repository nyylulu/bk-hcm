### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询计划类型列表。

### URL

POST /api/v1/woa/metas/plan_types/list

### 输入参数

无

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      "预测内",
      "预测外"
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
| details | string array | 计划类型列表 |
