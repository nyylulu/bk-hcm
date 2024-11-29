### 描述

- 该接口提供版本：v1.6.2.0+。
- 该接口所需权限：
- 该接口功能描述：查询AWS Savings Plan节省的成本。

### URL

POST /api/v1/account/vendors/aws/savings_plans/saved_cost/query

### 输入参数

所有查询条件之间为与关系

| 参数名称                   | 参数类型     | 必选 | 描述              |
|------------------------|----------|----|-----------------|
| root_account_id        | string   | 否  | 一级账号ID，为空查全部    |
| main_account_ids       | []string | 否  | 二级账号ID列表，为空查全部  |
| main_account_cloud_ids | []string | 否  | 二级账号云ID列表，为空查全部 |
| product_ids            | []int64  | 否  | 运营产品ID列表，为空查全部  |
| year                   | uint     | 是  | 年               |
| month                  | uint     | 是  | 月               |
| start_day              | uint     | 是  | 起始日             |
| end_day                | uint     | 是  | 截止日             |
| page                   | object   | 是  | 分页参数            |

### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

```json
{
  "root_account_id": "12345",
  "main_account_cloud_ids": [
    "cloud_id_1",
    "cloud_id_2"
  ],
  "product_ids": [
    111,
    222
  ],
  "year": 2024,
  "month": 8,
  "start_day": 1,
  "end_day": 31,
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "date": {
    "count": 0,
    "details": [
      {
        "main_account_id": "xxx",
        "main_account_cloud_id": "xxx",
        "main_account_managers": [],
        "main_account_bak_managers": [],
        "product_id": 123,
        "sp_arn": "xxx",
        "sp_managers": [],
        "sp_bak_managers": [],
        "unblended_cost": "15.55",
        "sp_effective_cost": "10.11",
        "sp_saved_cost": "5.44",
        "sp_net_effective_cost": "10.55"
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
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

### details

| 参数名称                      | 参数类型     | 描述                                                                                      |
|---------------------------|----------|-----------------------------------------------------------------------------------------|
| main_account_id           | string   | 二级账号云D                                                                                  |
| main_account_cloud_id     | string   | 二级账号云ID                                                                                 |
| main_account_managers     | []string | 二级账号负责人                                                                                 |
| main_account_bak_managers | []string | 二级账号备份负责人                                                                               |
| product_id                | int64    | 运营产品ID                                                                                  |
| sp_arn                    | string   | sp标识                                                                                    |
| sp_managers               | []string | sp所属账号负责人                                                                               |
| sp_bak_managers           | []string | sp所属账号备份负责人                                                                             |
| unblended_cost            | string   | 对应云资源的未混合成本, 对应：sum(line_item_unblended_cost)                                           |
| sp_effective_cost         | string   | sp有效成本，对应：sum(savings_plan_savings_plan_effective_cost)                                 |
| sp_saved_cost             | string   | sp节省成本，对应：sum(line_item_unblended_cost) - sum(savings_plan_savings_plan_effective_cost) |
| sp_net_effective_cost     | string   | sp有效成本净值，对应：sum(savings_plan_net_savings_plan_effective_cost)                           |
