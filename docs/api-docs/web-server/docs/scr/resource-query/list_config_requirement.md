### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：无。
- 该接口功能描述：需求类型配置信息查询。

### URL

GET /api/v1/woa/config/find/config/requirement

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
    "count": 3,
    "info":[
      {
        "id": 1,
        "require_type": 1,
        "require_name": "常规项目"
      },
      {
        "id": 2,
        "require_type": 2,
        "require_name": "春节保障"
      }，
      {
        "id": 3,
        "require_type": 3,
        "require_name": "机房裁撤"
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

| 参数名称 | 参数类型       | 描述                 |
|---------|--------------|----------------------|
| info    | object array | 需求类型配置信息详情列表 |

#### data.info

| 参数名称       | 参数类型 | 描述          |
|--------------|---------|---------------|
| id	       | int	 | 需求类型实例ID，系统内部管理ID |
| require_type | int	 | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤 |
| require_name | string	 | 需求类型描述    |
