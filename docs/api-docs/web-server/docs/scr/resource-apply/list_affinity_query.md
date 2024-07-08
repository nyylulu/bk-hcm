### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：无。
- 该接口功能描述：亲和性列表查询。

### URL

POST /api/v1/woa/config/find/config/affinity

### 输入参数

| 参数名称        | 参数类型  | 必选 | 描述        |
|---------------|----------|------|------------|
| resource_type | string   | 是   | 资源类型。"IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器      |
| has_zone	    | bool	   | 是   | 是否指定可用区 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "resource_type": "IDCDVM",
  "has_zone": true
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
  "data":{
    "info":[
      {
        "level":"ANTI_NONE",
        "description":"无需求"
      },
      {
        "level":"ANTI_RACK",
        "description":"分机架"
      },
      {
        "level":"ANTI_MODULE",
        "description":"分Module"
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
| info    | object array | 亲和性可选列表。ANTI_NONE: 无要求, AntiRack: 分机架, AntiModule: 分Module, AntiCampus: 分Campus              |

#### data.info

| 参数名称      | 参数类型    | 描述        |
|-------------|-----------|--------------|
| level       | int       | 反亲和性级别   |
| description | string    | 反亲和性描述   |
