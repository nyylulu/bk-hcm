### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：机房裁撤。
- 该接口功能描述：删除裁撤主机信息。

### URL

DELETE /api/v1/woa/dissolve/recycled_host/update

### 输入参数

| 参数名称  | 参数类型         | 必选 | 描述                |
|-------|--------------|----|-------------------|
| ids   | string array | 是	 | 主机唯一标识数组，最大支持100个 |

### 调用示例

```json
{
  "ids": ["1", "2"]
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
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型      | 描述               |
|------------|-------------|--------------------|
| result     | bool        | 请求成功与否。true:请求成功；false请求失败 |
| code       | int         | 错误编码。 0表示success，>0表示失败错误  |
| message    | string      | 请求失败返回的错误信息 |
| permission | object      | 权限信息             |
| request_id | string      | 请求链ID             |
| data	     | object | 响应数据             |
