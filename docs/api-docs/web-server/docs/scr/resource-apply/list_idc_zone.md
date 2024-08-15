### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：IDC可用区配置信息查询。

### URL

POST /api/v1/woa/config/findmany/config/idc/zone

### 输入参数

| 参数名称           | 参数类型      | 必选 | 描述     |
|------------------|--------------|------|---------|
| cmdb_region_name | string array | 是   | 地域列表。若列表非空，则返回地域列表下的区域信息；若列表为空，则返回所有地域下的区域信息 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "cmdb_region_name": ["上海"]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data":{
    "count":3,
    "info":[
      {
        "id":1,
        "cmdb_region_name": "上海",
        "cmdb_zone_name": "上海-青浦",
        "cmdb_zone_id": 141
      },
      {
        "id":2,
        "cmdb_region_name": "上海",
        "cmdb_zone_name": "上海-宝信",
        "cmdb_zone_id": 154
      },
      {
        "id":3,
        "cmdb_cmdb_region_name": "上海",
        "cmdb_zone_name": "上海-奉贤",
        "cmdb_zone_id": 217
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

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 可用区配置信息详情列表           |

#### data.info

| 参数名称          | 参数类型   | 描述         |
|------------------|----------|--------------|
| id               | int      | 可用区配置信息实例ID，系统内部管理ID |
| cmdb_region_name | string   | 地域，英文     |
| cmdb_zone_name   | string   | 可用区名      |
| cmdb_zone_id     | int      | cmdb可用区ID  |
