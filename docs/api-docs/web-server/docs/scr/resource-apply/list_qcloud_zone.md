### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：qcloud可用区配置信息查询。

### URL

POST /api/v1/woa/config/findmany/config/qcloud/zone

### 输入参数

| 参数名称          | 参数类型       | 必选 | 描述               |
|------------------|--------------|------|-------------------|
| region           | string	array | 否 	 | 资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器 |
| cmdb_region_name | string array | 否   | CMDB地域列表。若region为空而cmdb_region_name非空，则返回CMDB地域列表下的区域信息 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "region": ["ap-shanghai"]
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
    "count":2,
    "info":[
      {
        "id":1,
        "region":"ap-shanghai",
        "zone":"ap-shanghai-2",
        "zone_cn":"上海二区",
        "cmdb_zone_id":154,
        "cmdb_zone_name":"上海-宝信"
      },
      {
        "id":2,
        "region":"ap-shanghai",
        "zone":"ap-shanghai-3",
        "zone_cn":"上海三区",
        "cmdb_zone_id":187,
        "cmdb_zone_name":"上海-临港"
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
| info    | object array | 可用区配置信息详情列表      |

#### data.info

| 参数名称      | 参数类型  | 描述         |
|--------------|---------|--------------|
| id           | int	 | 可用区配置信息实例ID，系统内部管理ID      |
| region       | string  | 地域，英文     |
| zone	       | string  | 可用区，英文   |
| zone_cn      | string  | 可用区，中文   |
| cmdb_zone_id | string  | cmdb可用区ID  |
| zone_cn	   | string  | cmdb可用区名  |
