### 描述

- 该接口提供版本：v1.6.11+。
- 该接口所需权限：平台管理-滚服管理。
- 该接口功能描述：查询业务滚服基础额度和调整额度列表。

### URL

POST /api/v1/woa/rolling_servers/biz_quotas/list

### 输入参数

| 参数名称        | 参数类型         | 必选 | 描述                             |
|-------------|--------------|----|--------------------------------|
| bk_biz_ids  | int array    | 否  | 业务ID列表，不传时查询全部，数量最大100         |
| adjust_type | string array | 否  | 额度调整类型（枚举值：increase, decrease） |
| revisers    | string array | 否  | 更新人列表，精确匹配，数量最大100             |
| quota_month | string       | 是  | 额度所属月份，格式：YYYY-MM，例如：2024-09   |
| page        | object       | 是  | 分页设置                           |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                        |
|-------|--------|----|-------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start | int    | 否  | 记录开始位置，start 起始值为0                                                                                                                        |
| limit | int    | 否  | 每页限制条数，最大500，不能为0                                                                                                                         |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序，默认根据id正序排序，额外支持按照quota(基础额度)排序                                                                                         |
| order | string | 否  | 排序顺序，枚举值：ASC(升序)、DESC(降序)                                                                                                                 |

### 调用示例

```json
{
  "bk_biz_ids": [
    101
  ],
  "adjust_type": [
    "increase",
    "decrease"
  ],
  "reviser": [
    "Jim"
  ],
  "quota_month": "2024-09",
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
        "id": "00000011",
        "offset_config_id": "0000001d",
        "year": 2024,
        "month": 9,
        "bk_biz_id": 101,
        "bk_biz_name": "业务",
        "quota": 10000,
        "adjust_type": "increase",
        "quota_offset": 8000,
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2024-09-01T12:00:00Z",
        "updated_at": "2024-09-01T12:00:00Z"
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

| 参数名称             | 参数类型   | 描述                                                           |
|------------------|--------|--------------------------------------------------------------|
| id               | string | 业务滚服额度配置ID                                                   |
| offset_config_id | string | 业务滚服额度偏移配置ID（可用于查询偏移变更记录）。<br/>当业务滚服额度没有发生偏移时，该字段返回null      |
| year             | int    | 配置所属年份                                                       |
| month            | int    | 配置所属月份                                                       |
| bk_biz_id        | int    | 业务ID                                                         |
| bk_biz_name      | string | 业务名称                                                         |
| quota            | int    | 业务滚服基础额度                                                     |
| adjust_type      | string | 额度调整类型（枚举值：increase, decrease）。<br/>当业务滚服额度没有发生偏移时，该字段返回null |
| quota_offset     | int    | 业务滚服额度偏移量，绝对值，需要结合额度调整类型判断增减。<br/>当业务滚服额度没有发生偏移时，该字段返回null   |
| creator          | string | 创建人（偏移配置）                                                    |
| reviser          | string | 更新人（偏移配置）                                                    |
| created_at       | string | 创建时间（偏移配置），标准格式：2006-01-02T15:04:05Z                         |
| updated_at       | string | 更新时间（偏移配置），标准格式：2006-01-02T15:04:05Z                         |
