### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源回收子任务状态列表查询。

### URL

GET /api/v1/woa/task/find/config/recycle/status

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
        "description":"未执行",
        "status":"INIT"
      },
      {
        "description":"执行中",
        "status":"RUNNING"
      },
      {
        "description":"成功",
        "status":"SUCCESS"
      },
      {
        "description":"失败",
        "status":"FAILED"
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
| info    | object array | 回收状态列表信息           |

#### data.info

| 参数名称     | 参数类型  | 描述               |
|-------------|---------|--------------------|
| status      |	string  | 资源回收子任务状态    |
| desciption  |	string	| 资源回收子任务状态描述 |
