### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：CPU和MEM等规格配置信息查询。

### URL

GET /api/v1/woa/config/find/config/cvm/devicerestrict

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
    "cpu":[
      1,
      2,
      4,
      8
    ],
    "mem":[
      1,
      4,
      8
    ],
    "disk":[
      50,
      100
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

| 参数名称 | 参数类型       | 描述                |
|---------|--------------|---------------------|
| cpu	  | int array	 | cpu规格列表，单位核    |
| mem	  | int array	 | 内存大小规格列表，单位G |
| disk	  | int array	 | 磁盘大小规格列表，单位G |
