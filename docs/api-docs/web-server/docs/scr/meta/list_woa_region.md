### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：查询地区/城市列表。

### URL

GET /api/v1/woa/meta/region/list

### 输入参数

无

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "details": [
      {
        "region_id": "ap-shanghai",
        "region_name": "上海"
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

| 参数名称    | 参数类型         | 描述      |
|---------|--------------|---------|
| details | object array | 地区/城市列表 |

#### data.details[i]

| 参数名称        | 参数类型   | 描述      |
|-------------|--------|---------|
| region_id   | string | 地区/城市ID |
| region_name | string | 地区/城市名称 |
