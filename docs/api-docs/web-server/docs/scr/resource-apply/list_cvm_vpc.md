### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：私有网络VPC列表查询。

### URL

POST /api/v1/woa/config/findmany/config/cvm/vpc

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述     |
|--------|---------|------|---------|
| region | string  | 是   | 地域     |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "region": "ap-shanghai"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data":{
    "count":1,
    "info":[
      {
        "vpc_id": "vpc-2x7lhtse",
        "vpc_name": "VPC-IEG-SH"
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
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 私有网络信息列表           |

#### data.info

| 参数名称   | 参数类型   | 描述        |
|-----------|----------|-------------|
| vpc_id    | string   | 私有网络ID   |
| vpc_name  | string   | 私有网络名称  |
