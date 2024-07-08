### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：业务-主机申领。
- 该接口功能描述：资源申请改单推荐。

### URL

POST /api/v1/woa/task/recommend/apply

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述          |
|--------------|----------|------|--------------|
| suborder_id  | string	  | 是	 | 资源申请子单号 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "suborder_id":"1001-1"
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
    "suborder_id":"1001-1",
    "replicas":10,
    "spec":{
      "region":"ap-shanghai",
      "zone":"ap-shanghai-2",
      "device_type":"S3.LARGE8",
      "image_id":"img-r5igp4bv",
      "disk_size":200,
      "disk_type":"CLOUD_PREMIUM",
      "network_type":"TENTHOUSAND",
      "vpc":"",
      "subnet":""
    }
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

| 参数名称     | 参数类型  | 描述        |
|-------------|---------|------------|
| suborder_id |	int	    | CC业务ID    |
| replicas	  | int	    | 需求资源数量 |
| spec	      | object  | 资源需求声明 |

#### data.spec
| 参数名称      | 参数类型    | 描述              |
|--------------|-----------|-------------------|
| region       | string	   | 地域               |
| zone         | string	   | 可用区             |
| device_type  | string	   | 机型               |
| image_id     | string    | 镜像ID             |
| disk_size    | int       | 数据盘磁盘大小，单位G |
| disk_type	   | string	   | 数据盘磁盘类型。"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘 |
| network_type | string	   | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆 |
| vpc	       | string    | 私有网络，默认为空    |
| subnet       | string    | 私有子网，默认为空    |
