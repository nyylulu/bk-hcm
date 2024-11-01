### 描述

- 该接口提供版本：v1.6.11+。
- 该接口所需权限：平台管理-滚服管理。
- 该接口功能描述：查询已经有基础额度配置的业务列表，不分页，返回全部数据。

### URL

POST /api/v1/woa/rolling_servers/exist_quota_bizs/list

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述                           |
|-------------|--------|----|------------------------------|
| quota_month | string | 是  | 额度所属月份，格式：YYYY-MM，例如：2024-09 |

### 调用示例

```json
{
  "quota_month": "2024-09"
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
        "bk_biz_id": 101,
        "bk_biz_name": "业务",
        "quota": 10000
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

| 参数名称   | 参数类型         | 描述      |
|--------|--------------|---------|
| detail | object array | 查询返回的数据 |

#### data.details[n]

| 参数名称        | 参数类型   | 描述         |
|-------------|--------|------------|
| id          | string | 业务滚服额度配置ID |
| bk_biz_id   | int    | 业务ID       |
| bk_biz_name | string | 业务名称       |
| quota       | int    | 业务滚服基础额度   |
