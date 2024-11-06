### 描述

- 该接口提供版本：v1.6.11+。
- 该接口所需权限：无。
- 该接口功能描述：查询资源池业务列表。

### URL

POST /api/v1/woa/metas/respool_bizs/list

### 输入参数

无

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "details": [
      {
        "id": "00000011",
        "bk_biz_id": 101,
        "bk_biz_name": "业务",
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

| 参数名称        | 参数类型   | 描述                             |
|-------------|--------|--------------------------------|
| id          | string | 业务滚服额度配置ID                     |
| bk_biz_id   | int    | 业务ID                           |
| bk_biz_name | string | 业务名称                           |
| creator     | string | 创建人                            |
| reviser     | string | 更新人                            |
| created_at  | string | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at  | string | 更新时间，标准格式：2006-01-02T15:04:05Z |
