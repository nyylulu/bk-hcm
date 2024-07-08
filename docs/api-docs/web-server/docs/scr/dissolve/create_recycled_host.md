### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：平台-机房裁撤管理。
- 该接口功能描述：创建裁撤主机信息。

### URL

POST /api/v1/woa/dissolve/recycled_host/create

### 输入参数

| 参数名称    | 参数类型          | 必选 | 描述                |
|---------|---------------|------|-------------------|
| hosts | object array	 | 是	  | 裁撤主机信息列表，单次最多100个 |

| 参数名称     | 参数类型    | 必选 | 描述        |
|----------|---------|------|-----------|
| asset_id | string	 | 是	  | 主机固资号     |
| inner_ip | string  | 是	  | 主机ip      |
| module   | string  | 是	  | 机器所在的裁撤模块 |

### 调用示例

```json
{
  "hosts": [
    {
      "asset_id": "TC123456",
      "inner_ip": "127.0.0.1",
      "module": "深圳-锦绣-M12"
    },
    {
      "asset_id": "TC123457",
      "inner_ip": "127.0.0.2",
      "module": "深圳-锦绣-M12"
    }
  ]
}
```

### 响应示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "permission":null,
  "request_id":"f5a6331d4bc2433587a63390c76ba7bf",
  "data":{
    "ids": ["1", "2"]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| permission | object       | 权限信息             |
| request_id | string       | 请求链ID             |
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型         | 描述       |
|------|--------------|----------|
| ids  | string array | 裁撤主机id数组 |
