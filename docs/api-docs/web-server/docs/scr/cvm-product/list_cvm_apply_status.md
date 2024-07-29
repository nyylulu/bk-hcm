### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：CVM生产状态列表查询。

### URL

GET /api/v1/woa/cvm/find/config/apply/status

### 输入参数

无

### 调用示例

#### 获取详细信息请求参数示例

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
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述              |
|---------|--------------|-------------------|
| info    | object array | 状态列表信息        |

#### data.info

| 参数名称     | 参数类型 | 描述             |
|------------|---------|------------------|
| status     | string  | 单据状态          |
| desciption | string  | 资源回收步骤描述   |
