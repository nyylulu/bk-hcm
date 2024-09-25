### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询规划产品列表。

### URL

POST /api/v1/woa/metas/plan_products/list

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
        "plan_product_id": 902,
        "plan_product_name": "规划产品1"
      },
      {
        "plan_product_id": 901,
        "plan_product_name": "规划产品2"
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
| details | object array | 规划产品列表 |

#### data.details[n]

| 参数名称              | 参数类型   | 描述     |
|-------------------|--------|--------|
| plan_product_id   | int    | 规划产品ID |
| plan_product_name | string | 规划产品名称 |
