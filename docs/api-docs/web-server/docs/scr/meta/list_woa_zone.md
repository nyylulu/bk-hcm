### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：查询可用区列表。

### URL

POST /api/v1/woa/meta/zone/list

### 输入参数

| 参数名称       | 参数类型         | 必选  | 描述                  |
|------------|--------------|-----|---------------------|
| region_ids | string array | 否   | 查询地区/城市ID列表，不传时查询全部 |

### 调用示例

```json
{
  "region_ids": [
    "ap-shanghai"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "zone_id": "ap-shanghai-2",
        "zone_name": "上海二区"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型         | 描述    |
|---------|--------------|-------|
| details | object array | 可用区列表 |

#### data.details[n]

| 参数名称      | 参数类型   | 描述    |
|-----------|--------|-------|
| zone_id   | string | 可用区ID |
| zone_name | string | 可用区名称 |
