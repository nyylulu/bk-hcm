### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：匹配资源池设备执行。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/commit/apply/pool/match

### 输入参数

| 参数名称        | 参数类型       | 必选 | 描述              |
|----------------|--------------|------|------------------|
| suborder_id    | string	    | 是   | 资源申请子单号      |
| spec	         | object array | 是   | 匹配需求声明        |

#### spec
| 参数名称         | 参数类型       | 必选 | 描述               |
|-----------------|---------------|------|-------------------|
| device_type     | string	array | 是   | 机型               |
| bk_cloud_region | string	array | 是   | 地域ID               |
| bk_cloud_zone   | string	array | 是   | 可用区ID             |
| image_id        | string array  | 否   | 镜像ID。QCLOUDCVM必填，用于指定重装操作系统 |
| os_type	      | string	array | 否   | 操作系统。IDCPM必填，用于指定重装操作系统    |
| replicas	      | int           | 是   | 待匹配资源数量，最大为500 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "suborder_id":"1001-1",
  "spec":[
    {
      "device_type":"CG1-10G",
      "region":"深圳",
      "zone":"深圳-光明",
      "image_id":"",
      "os_type":"XServer V16_64",
      "replicas":10
    }
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data": null
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

无
