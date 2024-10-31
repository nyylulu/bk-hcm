### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台管理-滚服管理。
- 该接口功能描述：调整业务滚服额度，当该月不存在调整记录时会创建新记录，否则更新原有记录。

### URL

PATCH /api/v1/woa/rolling_servers/quota_offsets/batch

### 输入参数

| 参数名称         | 参数类型      | 必选 | 描述                             |
|--------------|-----------|----|--------------------------------|
| bk_biz_ids   | int array | 是  | 业务ID列表，最大100                   |
| adjust_month | object    | 是  | 调整月份                           |
| adjust_type  | string    | 是  | 额度调整类型（枚举值：increase, decrease） |
| quota_offset | int       | 是  | 额度调整值（大于等于0）                   |

#### adjust_month

| 参数名称  | 参数类型   | 必选 | 描述                                     |
|-------|--------|----|----------------------------------------|
| start | string | 是  | 额度调整生效起始时间，格式：YYYY-MM，例如：2024-09       |
| end   | string | 是  | 额度调整生效结束时间（包含当月），格式：YYYY-MM，例如：2024-09 |

### 调用示例

```json
{
  "bk_biz_ids": [
    639
  ],
  "adjust_month": {
    "start": "2024-09",
    "end": "2024-09"
  },
  "adjust_type": "decrease",
  "quota_offset": 8000
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "ids": [
      "00000001"
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

| 参数名称 | 参数类型         | 描述         |
|------|--------------|------------|
| ids  | string array | 滚服额度偏移配置ID |
