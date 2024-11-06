### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询运营产品列表。

### URL

POST /api/v1/woa/metas/op_products/list

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
        "op_product_id": 902,
        "op_product_name": "运营产品1"
      },
      {
        "op_product_id": 901,
        "op_product_name": "运营产品2"
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
| details | object array | 运营产品列表 |

#### data.details[n]

| 参数名称            | 参数类型   | 描述     |
|-----------------|--------|--------|
| op_product_id   | int    | 运营产品ID |
| op_product_name | string | 运营产品名称 |
