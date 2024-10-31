### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台管理-滚服管理。
- 该接口功能描述：查询业务滚服额度偏移配置调整操作记录。

### URL

POST /api/v1/woa/rolling_servers/quota_offsets/adjust_records/list

### 输入参数

| 参数名称              | 参数类型         | 必选 | 描述                   |
|-------------------|--------------|----|----------------------|
| offset_config_ids | string array | 是  | 滚服额度偏移配置ID列表，数量最大100 |
| page              | object       | 是  | 分页设置                 |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                        |
|-------|--------|----|-------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start | int    | 否  | 记录开始位置，start 起始值为0                                                                                                                        |
| limit | int    | 否  | 每页限制条数，最大500，不能为0                                                                                                                         |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序，默认根据created_at(调整时间)倒序排序，枚举值为：created_at(更新时间) 、quota_offset(调整量)                                                      |
| order | string | 否  | 排序顺序，枚举值：ASC(升序)、DESC(降序)                                                                                                                 |

### 调用示例

```json
{
  "offset_config_ids": [
    "0000001d"
  ],
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
        "operator": "Jim",
        "adjust_type": "decrease",
        "quota_offset": 8000,
        "created_at": "2024-09-01T12:00:00Z"
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

| 参数名称             | 参数类型   | 描述                             |
|------------------|--------|--------------------------------|
| id               | string | 操作记录ID                         |
| offset_config_id | string | 业务滚服额度偏移配置ID                   |
| operator         | string | 操作人                            |
| adjust_type      | string | 额度调整类型（枚举值：increase, decrease） |
| quota_offset     | int    | 业务滚服额度偏移量，绝对值，需要结合额度调整类型判断增减   |
| created_at       | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
