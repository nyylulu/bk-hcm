### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：匹配设备查询。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/findmany/apply/match/device

### 输入参数

| 参数名称             | 参数类型       | 必选 | 描述               |
|---------------------|--------------|------|-------------------|
| resource_type       | string	     | 是	| 资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器 |
| ips                 | string array | 否   | ip列表             |
| spec	              | object       | 是	| 资源需求声明        |
| anti_affinity_level | string       | 否	| 反亲和策略，默认值为"ANTI_NONE"。 "ANTI_NONE": 无要求, "ANTI_CAMPUS": 分Campus, "ANTI_MODULE": 分Module, "ANTI_RACK": 分机架 |
| total_num	          | int	         | 是	| 申请单据资源需求总量 |
| pending_num         |	int          | 是	| 待匹配资源数量      |

#### spec
| 参数名称      | 参数类型    | 描述              |
|--------------|---------|-------------------|
| region       | string	array | 地域               |
| zone         | string	array | 可用区             |
| device_type  | string	array | 机型               |
| image        | string array | 镜像名              |
| os_type	   | string	array | 操作系统            |
| raid_type	   | string	array | RAID类型           |
| disk_type	   | string	array | 数据盘磁盘类型。"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘 |
| network_type | string	array | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆 |
| isp	       | string	array | 外网运营商          |
| instance_charge_type     | string  | 实例计费模式。(PREPAID：表示预付费，即包年包月、POSTPAID_BY_HOUR：表示后付费，即按量计费、CDHPAID：专用宿主机付费，即只对专用宿主机计费，不对专用宿主机上的实例计费。) |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "resource_type":"QCLOUDCVM",
  "ips":[

  ],
  "spec":{
    "device_type":[
      "IT5.8XLARGE128"
    ],
    "image":[

    ],
    "kernel":[

    ],
    "disk_type":[

    ],
    "region":[
      "上海"
    ],
    "zone":[
      "上海-宝信"
    ]
  },
  "anti_affinity_level":"ANTI_NONE",
  "total_num":1,
  "pending_num":1
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
    "count":2,
    "info":[
      {
        "bk_host_id":17,
        "asset_id":"TC000000000001",
        "ip":"10.0.0.1",
        "outer_ip":"",
        "isp":"",
        "device_type":"S3ne.4XLARGE64",
        "os_type":"TencentOS Server 2.2 (Final)",
        "region":"上海",
        "zone":"上海-宝信",
        "module":"上海-宝信-M19",
        "equipment":581458,
        "idc_unit":"上海腾讯宝信DC电信7号楼M2-3-4",
        "idc_logic_area":"通用bonding合作业务25G区11",
        "raid_type":"NORAID",
        "input_time":"2022-04-07 00:00:00",
        "match_score":0.9,
        "match_tag":true
      },
      {
        "bk_host_id":18,
        "asset_id":"TC000000000002",
        "ip":"10.0.0.2",
        "outer_ip":"",
        "isp":"",
        "device_type":"S3ne.4XLARGE64",
        "os_type":"TencentOS Server 2.2 (Final)",
        "region":"上海",
        "zone":"上海-宝信",
        "module":"上海-宝信-M19",
        "equipment":581458,
        "idc_unit":"上海腾讯宝信DC电信7号楼M2-3-4",
        "idc_logic_area":"通用bonding合作业务25G区11",
        "raid_type":"NORAID",
        "input_time":"2022-04-07 00:00:00",
        "match_score":0,
        "match_tag":false
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

| 参数名称        | 参数类型  | 描述         |
|----------------|---------|--------------|
| bk_host_id     | int	   | CC主机ID      |
| asset_id	     | string  | 设备固资号     |
| ip	         | string  | 设备IP        |
| outer_ip	     | string  | 设备外网IP     |
| isp	         | string  | 外网运营商     |
| device_type    | string  | 机型          |
| os_type	     | string  | 操作系统类型   |
| region	     | string  | 地域          |
| zone	         | string  | 园区          |
| module         | string  | 模块          |
| equipment	     | int     | 机架号        |
| idc_unit	     | string  | IDC单元       |
| idc_logic_area | string  | 逻辑区域       |
| raid_type	     | string  | RAID类型      |
| input_time     | string  | 入库时间       |
| match_score    | float   | 匹配得分       |
| match_tag	     | bool    | 是否匹配推荐    |
