### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：无。
- 该接口功能描述：查询期望交付时间对应的需求可用周范围及可用年月范围。

### URL

POST /api/v1/woa/plans/demands/available_times/get

| 参数名称        | 参数类型   | 必选 | 描述                                |
|-------------|--------|----|-----------------------------------|
| expect_time | string | 是  | 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01 |

### 调用示例

```json
{
  "expect_time": "2024-09-01"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "year_month_week": {
      "year": 2024,
      "month": 8,
      "week_of_month": 5
    },
    "date_range_in_week": {
      "start": "2024-08-26",
      "end": "2024-09-01"
    },
    "date_range_in_month": {
      "start": "2024-07-29",
      "end": "2024-09-01"
    }
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

| 参数名称                | 参数类型   | 描述       |
|---------------------|--------|----------|
| year_month_week     | object | 需求年月周    |
| date_range_in_week  | object | 需求年月周天范围 |
| date_range_in_month | object | 需求年月天范围  |

#### year_month_week

| 参数名称          | 参数类型 | 描述       |
|---------------|------|----------|
| year          | int  | 需求年      |
| month         | int  | 需求月      |
| week_of_month | int  | 需求月内的第几周 |

#### date_range_in_week & date_range_in_month

| 参数名称  | 参数类型   | 描述                                          |
|-------|--------|---------------------------------------------|
| start | string | 起始时间，不能晚于当前时间，格式为YYYY-MM-DD，例如2024-01-01    |
| end   | string | 结束时间，不能早于start时间，格式为YYYY-MM-DD，例如2024-01-01 |
