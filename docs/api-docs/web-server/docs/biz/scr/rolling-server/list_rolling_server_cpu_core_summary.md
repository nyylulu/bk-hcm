### 描述

- 该接口提供版本：v1.6.11+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询滚服已交付、已退还的CPU核心数概览信息。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/rolling_servers/cpu_core/summary

### 输入参数

| 参数名称       | 参数类型        | 必选    | 描述                              |
|---------------|---------------|---------|----------------------------------|
| start         | object        | 是      | 申请开始时间                       |
| end           | object        | 是      | 申请结束时间，时间间隔不能超过一个月   |
| bk_biz_ids    | int array     | 否      | 业务ID数组，数量最大限制100          |
| order_ids     | string array  | 否      | 订单号数组，数量最大限制100          |
| suborder_ids  | string array  | 否      | 子订单号数组，数量最大限制100         |
| returned_way  | string        | 否      | 退还方式(枚举值(crp:通过crp退还、resource_pool:通过转移到资源池退还)) |
| applied_type  | string        | 否      | 申请类型(枚举值(normal:普通申请、resource_pool:资源池申请、cvm_product:管理员cvm生产)) |
| require_type  | int           | 否      | 项目类型(6:滚服项目8:春保资源池)) |

#### start

| 参数名称  | 参数类型   | 必选  | 描述       |
|----------|----------|------|------------|
| year     | int      | 是   | 时间年份    |
| month    | int      | 是   | 时间月份    |
| day      | int      | 是   | 时间天      |

#### end

| 参数名称  | 参数类型   | 必选  | 描述       |
|----------|----------|------|------------|
| year     | int      | 是   | 时间年份    |
| month    | int      | 是   | 时间月份    |
| day      | int      | 是   | 时间天      |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "start": {
    "year": 2024,
    "month": 10,
    "day": 1
  },
  "end": {
    "year": 2024,
    "month": 10,
    "day": 1
  },
  "order_ids": ["xxxxxx"],
  "suborder_ids": ["xxxxxx"],
  "bk_biz_ids": [111,222],
  "returned_way": "crp",
  "applied_type": "cvm_product",
  "require_type": 6
}
```

### 响应示例

#### 获取详细信息返回结果示例

如查询业务ID为1的滚服CPU核心数响应示例。

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "details": {
        "sum_delivered_core": 100,
        "sum_returned_applied_core": 200
      }
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型   | 描述   |
|---------|-----------|--------|
| code    | int       | 状态码  |
| message | string    | 请求信息 |
| data    | object    | 响应数据 |

#### data

| 参数名称  | 参数类型 | 描述                    |
|---------|---------|-------------------------|
| details | object  | 查询返回的数据            |

#### data.details

| 参数名称                     | 参数类型   | 描述                     |
|-----------------------------|----------|--------------------------|
| sum_delivered_core          | int      | cpu已交付的总核心数         |
| sum_returned_applied_core   | int      | cpu已退还的总核心数         |
