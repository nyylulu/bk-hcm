### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：DVM下的镜像列表查询。

### URL

GET /api/v1/woa/config/find/config/dvm/image

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
  "data":{
    "info":[
      "hub.oa.com/library/tlinux1.2:v1.17",
      "hub.oa.com/library/tlinux2.2:v1.6"
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

| 参数名称 | 参数类型       | 描述    |
|---------|--------------|---------|
| info    | string array | 镜像列表 |
