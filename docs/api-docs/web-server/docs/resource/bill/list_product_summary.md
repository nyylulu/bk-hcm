### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：账单查看。
- 该接口功能描述：查看根据运营产品聚合的账单信息（内部版）。

### URL

POST /api/v1/account/bills/product_summarys/list

### 输入参数

| 参数名称           | 参数类型      | 必选 | 描述     |
|----------------|-----------|----|--------|
| bill_year      | int       | 是  | 账单年份   |
| bill_month     | int       | 是  | 账单月份   |
| op_product_ids | int array | 是  | 运营产品id |
| page           | object    | 是  | 分页设置   |


#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |


### 调用示例

#### 获取详细信息请求参数示例

查看2024年6月, 运营id为2005000002的账单

```json
{
  "bill_year": 2024,
  "bill_month": 6,
  "op_product_ids": [2005000002],
  "page": {
    "limit": 10,
    "start": 0,
    "sort": "current_month_rmb_cost",
    "order": "DESC",
    "count": false
  }
}
```



### 响应示例

#### 导出成功结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 0,
    "details": [
      {
        "product_id": "2005000002",
        "product_name": "xxxx",
        "last_month_cost_synced": "0",
        "last_month_rmb_cost_synced": "0",
        "current_month_cost_synced": "0",
        "current_month_rmb_cost_synced": "0",
        "current_month_cost": "481878.27942169",
        "current_month_rmb_cost": "3429913.2172677051",
        "adjustment_cost": "0",
        "adjustment_rmb_cost": "0"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | string | 响应数据 |

#### data

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| count   | int32  | 总记录条数 |
| details | array  | 账单信息列表 |

#### details

| 参数名称                          | 参数类型   | 描述              |
|-------------------------------|--------|-----------------|
| product_id                    | string | 运营产品id          |
| product_name                  | string | 运营产品名称          |
| last_month_cost_synced        | string | 上月同步的账单总金额      |
| last_month_rmb_cost_synced    | string | 上月同步的账单总金额（人民币） |
| current_month_cost_synced     | string | 本月同步的账单总金额      |
| current_month_rmb_cost_synced | string | 本月同步的账单总金额（人民币） |
| current_month_cost            | string | 本月账单总金额         |
| current_month_rmb_cost        | string | 本月账单总金额（人民币）    |
| adjustment_cost               | string | 调整金额            |
| adjustment_rmb_cost           | string | 调整金额（人民币）       |

