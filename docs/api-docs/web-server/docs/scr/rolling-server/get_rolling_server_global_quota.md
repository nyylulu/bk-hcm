### 描述

- 该接口提供版本：v1.6.11+。
- 该接口所需权限：平台管理-滚服管理。
- 该接口功能描述：查询滚服全局额度配置。

### URL

GET /api/v1/woa/rolling_servers/global_config

### 输入参数

无

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "00000001",
    "global_quota": 200000,
    "biz_quota": 10000,
    "unit_price": 12.456,
    "creator": "Jim",
    "reviser": "Jim",
    "created_at": "2024-09-01T12:00:00Z",
    "updated_at": "2024-09-01T12:00:00Z"
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

| 参数名称         | 参数类型    | 描述                             |
|--------------|---------|--------------------------------|
| id           | string  | 全局配置ID                         |
| global_quota | int     | CPU总配额                         |
| biz_quota    | int     | 单业务CPU基础配额                     |
| unit_price   | decimal | 核算单价（核/天）                      |
| creator      | string  | 创建人                            |
| reviser      | string  | 更新人                            |
| created_at   | string  | 创建时间，标准格式：2006-01-02T15:04:05Z |
| updated_at   | string  | 更新时间，标准格式：2006-01-02T15:04:05Z |

