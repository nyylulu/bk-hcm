### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源回收单据重试。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/start/recycle/order

### 输入参数

| 参数名称                    | 参数类型         | 必选 | 描述     |
|-------------------------|--------------|----|--------|
| order_id                | []int        | 是  | 单据ID   |
| suborder_id             | []string     | 是  | 子单据ID  |
| return_forecast_configs | object array | 否  | 返还预测配置 |

ps: 
1.order_id 和 sub_order_id 只能同时传入一组。

#### return_forecast_configs

| 参数名称                 | 参数类型   | 必选 | 描述                                        |
|----------------------|--------|----|-------------------------------------------|
| suborder_id          | string | 是  | 回收子单据ID                                   |
| return_forecast      | bool   | 否  | 是否返还预测                                    |
| return_forecast_time | string | 否  | 期望返回预测时间，不能早于当天/不能晚于当年最后一天，格式是：YYYY-MM-DD |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id":[1001],
  "return_forecast_configs": [
    {
      "suborder_id": "123",
      "return_forecast": true,
      "return_forecast_time": "2025-09-17"
    }
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object array | 请求返回的数据        |

#### data

无
