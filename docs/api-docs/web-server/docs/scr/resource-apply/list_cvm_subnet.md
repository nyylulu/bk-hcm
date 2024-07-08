### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：无。
- 该接口功能描述：私有子网subnet列表查询。

### URL

POST /api/v1/woa/config/findmany/config/cvm/subnet

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述     |
|--------|---------|------|---------|
| region | string  | 是   | 地域     |
| zone   | string  | 是   | 可用区   |
| vpc    | string  | 是   | 私有网络  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "region": "ap-shanghai",
  "zone": "ap-shanghai-2",
  "vpc": "vpc-2x7lhtse"
}
```

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
        "subnet_id": "subnet-ax907buf",
        "subnet_name": "pass_use_0"
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
| info    | object array | 私有子网信息列表           |

#### data.info

| 参数名称      | 参数类型   | 描述        |
|--------------|----------|-------------|
| subnet_id    | string   | 私有子网ID   |
| subnet_name  | string   | 私有子网名称  |
