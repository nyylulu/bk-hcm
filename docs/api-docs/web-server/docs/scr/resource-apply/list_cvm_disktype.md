### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：无。
- 该接口功能描述：数据盘类型列表查询。

### URL

GET /api/v1/woa/config/find/config/cvm/disktype

### 输入参数

无

### 调用示例

无

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "permission":null,
  "request_id":"f5a6331d4bc2433587a63390c76ba7bf",
  "data":{
    "count":1,
    "info":[
      {
        "disk_type": "CLOUD_SSD",
        "disk_name": "SSD云硬盘"
      },
      {
        "disk_type": "CLOUD_PREMIUM",
        "disk_name": "高性能云盘"
      }
    ]
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

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 数据盘类型信息列表         |

#### data.info

| 参数名称   | 参数类型   | 描述          |
|-----------|----------|---------------|
| disk_type | string   | 数据盘类型     |
| disk_name | string   | 数据盘类型描述  |
