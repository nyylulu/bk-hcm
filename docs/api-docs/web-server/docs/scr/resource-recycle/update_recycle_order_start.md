### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：业务-主机回收。
- 该接口功能描述：资源回收单据重试。

### URL

POST /api/v1/woa/task/start/recycle/order

### 输入参数

| 参数名称   | 参数类型  | 必选 | 描述   |
|----------|----------|------|-------|
| order_id | int      | 是   | 单据ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id":1001
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
| data	     | object array | 请求返回的数据        |

#### data

无
