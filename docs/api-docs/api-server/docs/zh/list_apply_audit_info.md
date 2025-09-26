### 描述

- 该接口提供版本：v1.8.6+。
- 该接口所需权限：无。
- 该接口功能描述：查询主机申请单据审批信息。

### URL

POST /api/v1/woa/task/apply/ticket/audit/info/list

### 输入参数

| 参数名称       | 参数类型      | 必选 | 描述                |
|------------|-----------|----|-------------------|
| ticket_ids | int array | 是  | 申请单据id数组，最大长度为100 |


### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "ticket_ids": [38829]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "details": [
      {
        "ticket_id": 38829,
        "status": "RUNNING",
        "end_at": "",
        "current_steps":[
          {
            "name": "审核意见",
            "processors": ["admin"],
            "state_id": 8
          }
        ],
        "ticket_info":{
          "order_id":38829,
          "stage":"UNCOMMIT",
          "bk_biz_id":3,
          "bk_username":"xx",
          "follower":"",
          "enable_notice":true,
          "require_type":1,
          "expect_time":"2022-05-01 20:00:00",
          "remark":"",
          "suborders":[
            {
              "resource_type":"QCLOUDCVM",
              "replicas":2,
              "anti_affinity_level":"ANTI_NONE",
              "remark":"",
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
                "system_disk": {
                  "disk_type": "CLOUD_PREMIUM",
                  "disk_size": 100,
                  "disk_num": 1,
                },
                "data_disk": [{
                  "disk_type": "CLOUD_PREMIUM",
                  "disk_size": 100,
                  "disk_num": 1
                }]
              }
            },
            {
              "resource_type":"IDCPM",
              "replicas":2,
              "anti_affinity_level":"ANTI_NONE",
              "remark":"",
              "spec":{
                "region":"东莞",
                "zone":"东莞-大朗",
                "device_type":"B7",
                "os_type":"XServer V16_64",
                "raid_type":"RAID1",
                "network_type":"TENTHOUSAND",
                "isp":""
              }
            }
          ]
        }
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 响应数据                       |

#### data

| 参数名称    | 参数类型         | 描述         |
|---------|--------------|------------|
| details | object array | 资源申请单据审批信息 |

#### details

| 参数名称          | 参数类型         | 描述                                                                                |
|---------------|--------------|-----------------------------------------------------------------------------------|
| ticket_id     | string       | 申请单据id                                                                            |
| status        | string       | 资源申请单据审批状态，RUNNING(处理中)、FINISHED(已结束)、TERMINATED(被终止)、SUSPENDED(被挂起)、REVOKED(被撤销) |
| end_at        | time         | 当申请单据审批状态不为RUNNING时，返回该值，代表单据审批结束时间                                               |
| current_steps | object array | 当前步骤                                                                              |
| ticket_info   | object       | 资源申请单据详细信息                                                                        |

#### current_steps[0]
| 参数名称       | 参数类型         | 描述        |
|------------|--------------|-----------|
| name       | string	      | 步骤名称      |
| processors | string array | 处理人列表     |
| state_id   | int	         | 节点ID      |


#### ticket_info

| 参数名称           | 参数类型         | 描述                                                                |
|----------------|--------------|-------------------------------------------------------------------|
| order_id       | int	         | 若order_id传值且非0，则更新order_id对应的申请单据草稿；若order_id未传值或为0，则创建申请单据草稿     |
| stage          | string	      | 单据执行阶段。"UNCOMMIT": 未提交, "AUDIT": 审核中, "RUNNING": 生产中, "DONE": 已完成 |
| bk_biz_id      | int	         | CC业务ID                                                            |
| bk_username    | string       | 资源申请提单人                                                           |
| follower	      | string       | 关注人，如果有多人，以","分隔，如："name1,name2"                                  |
| enable_notice	 | boo          | 是否通知用户单据完成，默认为false                                               |
| require_type   | int	         | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤                                    |
| expect_time    | string       | 期望交付时间                                                            |
| remark	        | string       | 备注                                                                |
| suborders	     | object array | 资源申请子需求单信息                                                        |

#### ticket_info.suborders
| 参数名称                | 参数类型     | 描述                                                                                                        |
|---------------------|----------|-----------------------------------------------------------------------------------------------------------|
| resource_type	      | string   | 需求资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器                     |
| replicas		          | int	     | 需求资源数量                                                                                                    |
| anti_affinity_level | string   | 反亲和策略，默认值为"ANTI_NONE"。 "ANTI_NONE": 无要求, "ANTI_CAMPUS": 分Campus, "ANTI_MODULE": 分Module, "ANTI_RACK": 分机架 |
| remark	             | string   | 备注                                                                                                        |
| spec	               | object   | 资源需求声明                                                                                                    |

#### ticket_info.suborders.spec
| 参数名称         | 参数类型             | 描述                                                                       |
|--------------|------------------|--------------------------------------------------------------------------|
| region       | string	          | 地域                                                                       |
| zone         | string	          | 可用区                                                                      |
| device_group | string           | 机型类别                                                                     |
| device_type  | string	          | 机型                                                                       |
| image_id     | string           | 镜像ID                                                                     |
| disk_size    | int              | 数据盘磁盘大小，单位G（已废弃，优先使用data_disk字段）                                         |
| disk_type	   | string	          | 数据盘磁盘类型。"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘（已废弃，优先使用data_disk字段） |
| network_type | string	          | 网络类型。"ONETHOUSAND": 千兆, "TENTHOUSAND": 万兆                                |
| vpc	         | string           | 私有网络，默认为空                                                                |
| subnet       | string           | 私有子网，默认为空                                                                |
| os_type      | string	          | 操作系统                                                                     |
| raid_type    | string	          | RAID类型                                                                   |
| isp          | string	          | 外网运营商                                                                    |
| mount_path   | string	          | 数据盘挂载点                                                                   |
| cpu_provider | string	          | CPU类型                                                                    |
| kernel       | string           | 内核                                                                       |
| system_disk  | DiskObject       | 系统盘，磁盘大小：50G-1000G且为50的倍数（IT类型默认本地盘、50G；其他类型默认高性能云盘、100G）                |
| data_disk    | array DiskObject | 数据盘，支持多块硬盘，磁盘大小：20G-32000G且为10的倍数，数据盘数量总和不能超过20块                         |

#### ticket_info.suborders.spec.DiskObject
| 参数名称   | 参数类型  | 描述                                                      |
|-----------|-------|----------------------------------------------------------|
| disk_type | string | 磁盘类型，"CLOUD_SSD": SSD云硬盘, "CLOUD_PREMIUM": 高性能云盘 |
| disk_size | int   | 磁盘大小，单位G                                             |
| disk_num  | int   | 磁盘数量，所有磁盘数量之和不能超过20块，默认1块                  |
