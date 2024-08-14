### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：资源申请改单记录查询。

### URL

POST /api/v1/woa/task/find/apply/record/modify

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述          |
|--------------|--------------|------|--------------|
| suborder_id  | string	array | 是	 | 资源申请子单号 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "suborder_id":["1001-1"]
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
        "id":1,
        "suborder_id":"1001-1",
        "bk_username":"xxx",
        "details":{
          "pre_data":{
            "replicas":4,
            "region":"ap-shanghai",
            "zone":"ap-shanghai-5",
            "device_type":"SA3.4XLARGE64",
            "image_id":"img-evitcbqz",
            "disk_size":300,
            "disk_type":"CLOUD_SSD",
            "network_type":"TENTHOUSAND",
            "vpc":"",
            "subnet":""
          },
          "cur_data":{
            "replicas":4,
            "region":"ap-shanghai",
            "zone":"ap-shanghai-4",
            "device_type":"SA3.4XLARGE64",
            "image_id":"img-evitcbqz",
            "disk_size":300,
            "disk_type":"CLOUD_SSD",
            "network_type":"TENTHOUSAND",
            "vpc":"",
            "subnet":""
          }
        },
        "create_at":"2022-10-15 18:07:37"
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

| 参数名称  | 参数类型       | 描述              |
|----------|--------------|-------------------|
| count    | int	      | 记录条数           |
| info	   | object array | 资源申请改单记录列表 |

#### data.info
| 参数名称      | 参数类型   | 描述           |
|-------------|-----------|----------------|
| id	      | int	      | 记录ID         |
| suborder_id | string	  | 资源申请子单号   |
| bk_username | string	  | 改单操作人      |
| details	  | object	  | 改单详情        |
| create_at	  | timestamp | 记录创建时间     |

#### data.info.details
| 参数名称     | 参数类型   | 描述            |
|-------------|----------|-----------------|
| pre_data	  | object	 | 变更前申请单据信息 |
| cur_data	  | object	 | 变更后申请单据信息 |

#### data.info.details.pre_data && cur_data
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
