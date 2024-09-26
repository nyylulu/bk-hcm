### 描述

- 该接口提供版本：v1.6.2+。
- 该接口所需权限：平台-CVM生产。
- 该接口功能描述：创建CVM生产单据。

### URL

POST /api/v1/woa/cvm/create/apply/order

### 输入参数

| 参数名称       | 参数类型       | 必选 | 描述             |
|---------------|--------------|------|-----------------|
| bk_biz_id     | int	       | 是	  | CC业务ID         |
| bk_module_id  | int          | 是	  | CC业务模块ID      |
| bk_username   | string       | 是	  | 资源申请提单人     |
| require_type  | int	       | 是	  | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤 |
| replicas      | int          | 是	  | 需求资源数量       |
| remark	    | string       | 否	  | 备注              |
| spec	        | object       | 是   | 资源需求声明        |

#### spec
| 参数名称       | 参数类型    | 必选 | 描述                                                  |
|---------------|---------|----|-----------------------------------------------------|
| region        | string	 | 是  | 地域                                                  |
| zone          | string	 | 是  | 可用区                                                 |
| device_type   | string	 | 是  | 机型                                                  |
| image_id      | string  | 是  | 镜像ID                                                |
| disk_size     | int     | 是  | 数据盘磁盘大小，单位G                                         |
| disk_type	    | string	 | 是  | 数据盘磁盘类型。"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘 |
| network_type  | string	 | 是  | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆           |
| vpc	        | string  | 否  | 私有网络，默认为空                                           |
| subnet        | string  | 否  | 私有子网，默认为空                                           |
| charge_type   | string  | 是  | 计费模式 (PREPAID:包年包月，POSTPAID_BY_HOUR:按量计费)           |
| charge_months | int     | 否  | 计费时长，单位：月(计费模式为包年包月时，该字段必传)                         |
| inherit_instance_id | string  | 否  | 被继承云主机实例ID（同一批次只支持一台），如果是滚服项目，该字段必传                 |

### 调用示例

#### 获取详细信息请求参数示例

#### CVM申请示例
```json
{
  "bk_biz_id":931,
  "bk_module_id": 51524,
  "bk_username":"xx",
  "require_type":1,
  "remark":"",
  "replicas":2,
  "spec":{
    "region":"ap-shanghai",
    "zone":"ap-shanghai-2",
    "device_type":"S3.LARGE8",
    "image_id":"img-r5igp4bv",
    "disk_size":200,
    "disk_type":"CLOUD_PREMIUM",
    "network_type":"TENTHOUSAND",
    "vpc":"",
    "subnet":"",
    "charge_type":"PREPAID",
    "charge_months":1,
    "inherit_instance_id": "ins-111"
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
    "order_id": 1001
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

| 参数名称  | 参数类型 | 描述   |
|----------|--------|--------|
| order_id | int    | 单据ID |
