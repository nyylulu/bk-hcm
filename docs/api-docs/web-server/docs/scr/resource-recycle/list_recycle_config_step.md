### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：无。
- 该接口功能描述：资源回收步骤配置列表查询。

### URL

GET /api/v1/woa/task/find/config/recycle/step

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
    "info":[
      {
        "id":1,
        "sequence":1,
        "name":"PRE_CHECK",
        "description":"检查CC模块和负责人",
        "retry":5
      },
      {
        "id":2,
        "sequence":2,
        "name":"CHECK_UWORK",
        "description":"检查是否有Uwork故障单据",
        "retry":5
      },
      {
        "id":3,
        "sequence":3,
        "name":"CHECK_GCS",
        "description":"检查是否有GCS记录",
        "retry":5
      },
      {
        "id":4,
        "sequence":4,
        "name":"BASIC_CHECK",
        "description":"tmp,tgw,tgw nat,l5策略检查",
        "retry":5
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
| info    | object array | 回收步骤配置列表信息        |

#### data.info

| 参数名称     | 参数类型 | 描述                 |
|------------|---------|----------------------|
| id         |	int    | 资源回收步骤配置ID      |
| sequence   | int	   | 资源回收步骤序号        |
| name       | string  | 资源回收步骤名          |
| desciption | string  | 资源回收步骤描述        |
| retry      | int     | 资源回收步骤最大重试次数 |
