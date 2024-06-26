### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：IaaS资源操作。
- 该接口功能描述：资源申请单据终止。

### URL

POST /api/v1/woa/task/terminate/apply

### 输入参数

| 参数名称     | 参数类型       | 必选 | 描述                    |
|-------------|--------------|------|------------------------|
| suborder_id | string array | 是   | 资源申请子单号，数量最大20 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "suborder_id":["1-1"]
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
  "data": null
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

无
