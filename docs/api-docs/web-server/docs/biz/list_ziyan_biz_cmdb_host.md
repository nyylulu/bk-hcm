### 描述

- 该接口提供版本：v1.4.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询CC业务下的自研云主机。注意，query_from_cloud
  为false时仅返回cloud_id,region,ip信息。该接口返回的所有本地id字段字段为空，请使用cloud_id。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/vendors/tcloud-ziyan/cmdb/hosts/list

### 输入参数

| 参数名称             | 参数类型      | 必选 | 描述        |
|------------------|-----------|----|-----------|
| bk_biz_id        | int       | 是  | 业务ID      |
| query_from_cloud | bool      | 否  | 是否从云上拉取数据 |
| account_id       | string    | 是  | 账号id      |
| region           | string    | 否  | 地域        |
| bk_set_ids       | array int | 否  | 集群ID列表    |
| bk_module_ids    | array int | 否  | 模块ID列表    |
| page             | object    | 是  | 分页设置      |

#### page

| 字段    | 类型     | 必选 | 描述               |
|-------|--------|----|------------------|
| start | int    | 是  | 记录开始位置           |
| limit | int    | 是  | 每页限制条数,**最大100** |
| sort  | string | 否  | 排序字段             |

### 调用示例

查询业务100下广州地域的，集群ID为1下的，模块ID为2，3的，云主机列表。

```json
{
  "region": "ap-guangzhou",
  "bk_set_ids": [
    1
  ],
  "bk_module_ids": [
    2,
    3
  ],
  "page": {
    "start": 0,
    "limit": 100
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 2,
    "details": [
      {
        "id": "",
        "cloud_id": "cvm-123",
        "name": "cvm-test",
        "vendor": "tcloud",
        "bk_biz_id": -1,
        "bk_cloud_id": 100,
        "account_id": "0000001",
        "region": "ap-hk",
        "zone": "ap-hk-1",
        "cloud_vpc_ids": [
          "vpc-123"
        ],
        "cloud_subnet_ids": [
          "subnet-123"
        ],
        "cloud_image_id": "image-123",
        "os_name": "linux",
        "memo": "cvm test",
        "status": "init",
        "private_ipv4_addresses": [
          "127.0.0.1"
        ],
        "private_ipv6_addresses": [],
        "public_ipv4_addresses": [
          "127.0.0.2"
        ],
        "public_ipv6_addresses": [],
        "machine_type": "s5",
        "cloud_created_time": "2022-01-20",
        "cloud_launched_time": "2022-01-21",
        "cloud_expired_time": "2022-02-22",
        "extension": {
          "cloud_security_group_ids": ["sg-111"],
          "security_group_names": ["sg1"]
        },
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2023-02-12T14:47:39Z",
        "updated_at": "2023-02-12T14:55:40Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称    | 参数类型   | 描述             |
|---------|--------|----------------|
| count   | uint64 | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据        |

#### data.details[n]

| 参数名称                   | 参数类型                     | 描述                                   |
|------------------------|--------------------------|--------------------------------------|
| ~~id~~                 | string                   | ~~资源ID~~  **为空**                     |
| cloud_id               | string                   | 云资源ID                                |
| name                   | string                   | 名称                                   |
| vendor                 | string                   | 供应商（枚举值：tcloud-ziyan）                |
| bk_biz_id              | int64                    | 业务ID                                 |
| bk_cloud_id            | int64                    | 云区域ID                                |
| account_id             | string                   | 账号ID                                 |
| region                 | string                   | 地域                                   |
| zone                   | string                   | 可用区                                  |
| cloud_vpc_ids          | string array             | 云VpcID列表                             |
| cloud_subnet_ids       | string array             | 云子网ID列表                              |
| cloud_image_id         | string                   | 云镜像ID                                |
| os_name                | string                   | 操作系统名称                               |
| memo                   | string                   | 备注                                   |
| status                 | string                   | 状态                                   |
| private_ipv4_addresses | string array             | 内网IPv4地址                             |
| private_ipv6_addresses | string array             | 内网IPv6地址                             |
| public_ipv4_addresses  | string array             | 公网IPv4地址                             |
| public_ipv6_addresses  | string array             | 公网IPv6地址                             |
| machine_type           | string                   | 设备类型                                 |
| cloud_created_time     | string                   | Cvm在云上创建时间，标准格式：2006-01-02T15:04:05Z |
| cloud_launched_time    | string                   | Cvm启动时间，标准格式：2006-01-02T15:04:05Z    |
| cloud_expired_time     | string                   | Cvm过期时间，标准格式：2006-01-02T15:04:05Z    |
| extension              | object[tcloud_extension] | 混合云差异字段                              |                          
| creator                | string                   | 创建者                                  |
| reviser                | string                   | 修改者                                  |
| created_at             | string                   | 创建时间，标准格式：2006-01-02T15:04:05Z       |
| updated_at             | string                   | 修改时间，标准格式：2006-01-02T15:04:05Z       |

                            |

#### tcloud_extension

| 参数名称                     | 参数类型                      | 描述                                                                                                                                                                |
|--------------------------|---------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| placement                | TCloudPlacement           | 位置信息。                                                                                                                                                             |
| instance_charge_type     | string                    | 实例计费模式。(PREPAID：表示预付费，即包年包月、POSTPAID_BY_HOUR：表示后付费，即按量计费、CDHPAID：专用宿主机付费，即只对专用宿主机计费，不对专用宿主机上的实例计费。、SPOTPAID：表示竞价实例付费。)。                                           |
| cpu                      | int64                     | Cpu。                                                                                                                                                              |
| memory                   | int64                     | 内存。                                                                                                                                                               |
| cloud_system_disk_id     | string                    | 云系统硬盘ID。                                                                                                                                                          |
| cloud_data_disk_ids      | string array              | 云数据盘ID。                                                                                                                                                           |
| internet_accessible      | TCloudInternetAccessible  | 描述了实例的公网可访问性，声明了实例的公网使用计费模式，最大带宽等。                                                                                                                                |
| virtual_private_cloud    | TCloudVirtualPrivateCloud | 描述了网络信息等。                                                                                                                                                         |
| renew_flag               | string                    | 自动续费标识。注意：后付费模式本项为null。取值范围：- NOTIFY_AND_MANUAL_RENEW：表示通知即将过期，但不自动续费 - NOTIFY_AND_AUTO_RENEW：表示通知即将过期，而且自动续费 - DISABLE_NOTIFY_AND_MANUAL_RENEW：表示不通知即将过期，也不自动续费。 |
| cloud_security_group_ids | string array              | 云安全组ID。                                                                                                                                                           |
| stop_charging_mode       | string                    | 实例的关机计费模式。取值范围：- KEEP_CHARGING：关机继续收费- STOP_CHARGING：关机停止收费- NOT_APPLICABLE：实例处于非关机状态或者不适用关机停止计费的条件。                                                              |
| uuid                     | string                    | 云UUID。                                                                                                                                                            |
| isolated_source          | string                    | 实例隔离类型。取值范围：- ARREAR：表示欠费隔离XPIRE：表示到期隔离ANMADE：表示主动退还隔离OTISOLATED：表示未隔离。                                                                                           |
| disable_api_termination  | bool                      | 实例销毁保护标志，表示是否允许通过api接口删除实例。默认取值：FALSE。取值范围：- TRUE：表示开启实例保护，不允许通过api接口删除实例ALSE：表示关闭实例保护，允许通过api接口删除实例                                                              |
| security_group_names     | string array              | 安全组名称顺序和安全云id一致                                                                                                                                                   |