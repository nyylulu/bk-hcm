### 描述

- 该接口提供版本：v1.7.1.0+。
- 该接口所需权限：业务-资源预测操作。
- 该接口功能描述：批量取消资源预测需求。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/demands/cancel

### 输入参数

| 参数名称           | 参数类型         | 必选 | 描述                |
|----------------|--------------|----|-------------------|
| cancel_demands | object array | 是  | 删除的预测需求列表，数量最大100 |

#### cancel_demands

| 参数名称              | 参数类型   | 必选 | 描述           |
|-------------------|--------|----|--------------|
| demand_id         | string | 是  | 预测需求ID       |
| remained_cpu_core | int64  | 是  | 预测需求剩余的CPU数量 |

### 调用示例

```json
{
  "cancel_demands": [
    {
      "demand_id": "0000001z",
      "remained_cpu_core": 100
    },
    {
      "demand_id": "0000002a",
      "remained_cpu_core": 24
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "00000001"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data	   | object | 响应数据                      |

#### data

| 参数名称 | 参数类型   | 描述     |
|------|--------|--------|
| id   | string | 预测单据ID |
