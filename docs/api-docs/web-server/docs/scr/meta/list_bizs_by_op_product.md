### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：无。
- 该接口功能描述：根据运营产品ID查询业务列表。

### URL

POST /api/v1/woa/metas/bizs/by/op_product/list

### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述     |
|---------------|------|----|--------|
| op_product_id | int  | 是  | 运营产品ID |

### 调用示例

```json
{
  "op_product_id": 902
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "bk_biz_id": 639,
        "bk_biz_name": "biz name"
      },
      {
        "bk_biz_id": 2,
        "bk_biz_name": "biz name2"
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

| 参数名称    | 参数类型         | 描述   |
|---------|--------------|------|
| details | object array | 业务列表 |

#### data.details[n]

| 参数名称        | 参数类型   | 描述   |
|-------------|--------|------|
| bk_biz_id   | int    | 业务ID |
| bk_biz_name | string | 业务名称 |