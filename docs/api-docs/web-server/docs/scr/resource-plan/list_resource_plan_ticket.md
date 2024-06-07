### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询资源预测单据。

### URL

POST /api/v1/woa/plan/resource/ticket/list

| 参数名称              | 参数类型         | 必选 | 描述                          |
|-------------------|--------------|----|-----------------------------|
| bk_biz_ids        | int array    | 否  | 业务ID列表，不传时查询全部              |
| ticket_ids        | string array | 否  | 资源预测需求单据ID列表，不传时查询全部，最多传20个 |
| applicants        | string array | 否  | 申请人列表，不传时查询全部，最多传20个        |
| submit_time_range | object       | 否  | 提单时间范围                      |
| page              | object       | 是  | 分页设置                        |

### submit_time_range

| 参数名称  | 参数类型   | 必选 | 描述                                          |
|-------|--------|----|---------------------------------------------|
| start | string | 是  | 起始时间，不能晚于当前时间，格式为YYYY-MM-DD，例如2024-01-01    |
| end   | string | 是  | 结束时间，不能早于start时间，格式为YYYY-MM-DD，例如2024-01-01 |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                                        |
|-------|--------|----|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count                                 |
| start | int    | 否  | 记录开始位置，start 起始值为0                                                                                                                                                        |
| limit | int    | 否  | 每页限制条数，最大500，不能为0                                                                                                                                                         |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序，默认根据submitted_at(提单时间)倒序排序，枚举值为：cpu_core(CPU核心数)、memory(内存大小)、disk_size(云盘大小)、expect_time(期望交付时间)、submitted_at(提单时间)、created_at(创建时间)、updated_at(更新时间) |
| order | string | 否  | 排序顺序，枚举值：ASC(升序)、DESC(降序)                                                                                                                                                 |

### 调用示例

```json
{
  "bk_biz_id": [
    639
  ],
  "ticket_ids": [
    "00000001"
  ],
  "applicants": [
    "shuotan"
  ],
  "submit_time_range": {
    "start": "2023-03-01",
    "end": "2023-06-01"
  },
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
  "message": "success",
  "data": {
    "details": [
      {
        "id": "00000001",
        "bk_biz_id": 111,
        "bk_biz_name": "业务",
        "bk_product_id": 222,
        "bk_product_name": "运营产品",
        "plan_product_id": 333,
        "plan_product_name": "规划产品",
        "demand_class": "CVM",
        "cpu_core": 123,
        "memory": 123,
        "disk_size": 123,
        "demand_week": "PLAN_0_4W",
        "demand_week_name": "0-4周",
        "remark": "这里是预测说明",
        "applicant": "tom",
        "submitted_at": "2019-07-29 11:57:20",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20"
      }
    ]
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

| 参数名称   | 参数类型         | 描述                                       |
|--------|--------------|------------------------------------------|
| count  | int          | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回 |
| detail | object array | 查询返回的数据，仅在 count 查询参数设置为 false 时返回       |

#### data.details[n]

| 参数名称              | 参数类型   | 描述               |
|-------------------|--------|------------------|
| id                | string | 资源预测需求单据ID       |
| bk_biz_id         | int    | 业务ID             |
| bk_biz_name       | string | 业务名称             |
| bk_product_id     | int    | 运营产品ID           |
| bk_product_name   | string | 运营产品名称           |
| plan_product_id   | int    | 规划产品ID           |
| plan_product_name | string | 规划产品名称           |
| demand_class      | string | 预测的需求类型          |
| cpu_core          | int    | 总CPU核心数，单位：核     |
| memory            | int    | 总内存大小，单位：GB      |
| disk_size         | int    | 总云盘大小，单位：GB      |
| demand_week       | string | 13周需求类型，由CRP系统定义 |
| demand_week_name  | string | 13周需求类型名称        |
| remark            | string | 预测说明             |
| applicant         | string | 申请人              |
| submitted_at      | string | 提单时间             |
| created_at        | string | 创建时间             |
| updated_at        | string | 更新时间             |
