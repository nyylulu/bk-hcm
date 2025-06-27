### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：根据回收单号查询回收设备列表。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/pool/findmany/recall/match/device

### 输入参数

| 参数名称       | 参数类型     | 必选 | 描述        |
|---------------|------------|------|------------|
| resource_type	| string	 | 是   | 资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机 |
| spec          | object     | 否   | 资源需求声明 |

#### spec
| 参数名称          | 参数类型       | 必选 | 描述        |
|------------------|--------------|------|------------|
| device_group     | string array | 否   | 机型类别     |
| bk_cloud_regions | string array | 否   | 地域ID数组   |
| bk_cloud_zones   | string array | 否   | 可用区ID数组 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "resource_type":"QCLOUDCVM",
  "spec":{
    "device_type":[
      "IT5.8XLARGE128"
    ],
    "bk_cloud_regions":[
      "ap-nanjing"
    ],
    "bk_cloud_zones":[
      "ap-nanjing-1"
    ]
  }
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
    "count":1,
    "info":[
      {
        "resource_type":"QCLOUDCVM",
        "spec":{
          "device_type": "IT5.8XLARGE128",
          "bk_cloud_region": "ap-nanjing",
          "bk_cloud_zone": "ap-nanjing-1"
        }
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
| info    | object array | 匹配设备列表              |

#### data.info

| 参数名称         | 参数类型  | 描述                     |
|-----------------|---------|--------------------------|
| device_type	  | string  | 机型                      |
| bk_cloud_region | string  | 地域ID，示例：ap-nanjing   |
| bk_cloud_zone	  | string  | 园区ID，示例：ap-nanjing-1 |
| amount	      | int	    | 数量                      |
