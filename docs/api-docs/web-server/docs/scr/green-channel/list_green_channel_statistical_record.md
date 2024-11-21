### 描述

- 该接口提供版本：v9.9.9。
- 该接口所需权限：平台管理-小额绿通。
- 该接口功能描述：查询小额绿通分析统计记录。

### URL

POST /api/v1/woa/green_channels/statistical_record/list

### 输入参数

| 参数名称       | 参数类型        | 必选    | 描述                      |
|---------------|---------------|---------|-------------------------|
| start         | object        | 是      | 申请开始时间                  |
| end           | object        | 是      | 申请结束时间                  |
| bk_biz_ids    | int array     | 否      | 业务ID数组，数量最大限制100，不传时为全选 |
| page          | object       | 是  | 分页设置                    |

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

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                        |
|-------|--------|----|-------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start | int    | 否  | 记录开始位置，start 起始值为0                                                                                                                        |
| limit | int    | 否  | 每页限制条数，最大500，不能为0                                                                                                                         |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序，默认根据created_at(调整时间)倒序排序，枚举值为：created_at(更新时间) 、quota_offset(调整量)                                                      |
| order | string | 否  | 排序顺序，枚举值：ASC(升序)、DESC(降序)                                                                                                                 |

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
  "bk_biz_ids": [111,222],
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "details": [
      {
        "bk_biz_id": 111,
        "order_count": 1,
        "sum_delivered_core": 1,
        "sum_applied_core": 1
      }
    ]  
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

| 参数名称   | 参数类型         | 描述                                       |
|--------|--------------|------------------------------------------|
| count  | int          | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回 |
| detail | object array | 查询返回的数据，仅在 count 查询参数设置为 false 时返回       |

#### data.details[n]
| 参数名称               | 参数类型   | 描述         |
|--------------------|----------|------------|
| bk_biz_id          | int      | 业务id       |
| order_count        | int      | 申请单据数量     |
| sum_delivered_core | int      | cpu已交付的总核心数 |
| sum_applied_core   | int      | cpu申请的总核心数 |
