### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务-资源预测操作。
- 该接口功能描述：创建资源预测单据。

### URL

POST /api/v1/woa/plan/resource/ticket/create

| 参数名称         | 参数类型         | 必选 | 描述                  |
|--------------|--------------|----|---------------------|
| bk_biz_id    | int          | 是  | 业务ID                |
| demand_class | string       | 是  | 预测的需求类型(枚举值：CVM、CA) |
| demands      | object array | 是  | 需求列表                |
| remark       | string       | 是  | 预测说明，最少20字，最多1024字  |

#### demands[i]

| 参数名称             | 参数类型         | 必选 | 描述                                                |
|------------------|--------------|----|---------------------------------------------------|
| obs_project      | string       | 是  | OBS项目类型                                           |
| expect_time      | string       | 是  | 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01                 |
| region_id        | string       | 是  | 地区/城市ID                                           |
| zone_id          | string       | 否  | 可用区ID                                             |
| demand_source    | string       | 是  | 需求分类/变更原因                                         |
| remark           | string       | 否  | 需求备注                                              |
| demand_res_types | string array | 是  | 预测资源类型列表(枚举值：CVM、CBS)，需求包含CVM时，传递CVM，包含CBS时，传递CBS |
| cvm              | object       | 否  | 申请的CVM信息                                          |
| cbs              | object       | 否  | 申请的CBS信息                                          |

#### demands[i].cvm

| 参数名称        | 参数类型   | 必选 | 描述                 |
|-------------|--------|----|--------------------|
| res_mode    | string | 是  | 资源模式(枚举值：按机型、按机型族) |
| device_type | string | 是  | 机型规格               |
| os          | int    | 是  | OS数，单位：台           |
| cpu_core    | int    | 是  | CPU核心数，单位：核        |
| memory      | int    | 是  | 内存大小，单位：GB         |

#### demands[i].cbs

| 参数名称      | 参数类型   | 必选 | 描述                                                |
|-----------|--------|----|---------------------------------------------------|
| disk_type | string | 是  | 云盘类型(枚举值：CLOUD_PREMIUM(高性能云硬盘)、CLOUD_SSD(SSD云硬盘)) |
| disk_io   | int    | 是  | 磁盘IO吞吐需求，无特殊要求填写15；高性能云盘上限150，SSD云硬盘上限260         |
| disk_size | int    | 是  | 云盘大小，单位：GB                                        |

### 调用示例

```json
{
  "bk_biz_id": 639,
  "demand_class": "CVM",
  "demands": [
    {
      "obs_project": "常规项目",
      "expect_time": "2024-11-12",
      "region_id": "ap-shanghai",
      "zone_id": "ap-shanghai-2",
      "demand_source": "指标变化",
      "remark": "这里是需求备注",
      "demand_res_types": [
        "CVM",
        "CBS"
      ],
      "cvm": {
        "res_mode": "按机型",
        "device_type": "S5.2XLARGE16",
        "os": 123,
        "cpu_core": 123,
        "memory": 123
      },
      "cbs": {
        "disk_type": "CLOUD_PREMIUM",
        "disk_io": 123,
        "disk_size": 1024
      }
    }
  ],
  "remark": "这是一个备注，这是一个备注，这是一个备注"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "00000001"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data	   | object | 响应数据                      |

#### data

| 参数名称 | 参数类型   | 描述     |
|------|--------|--------|
| id   | string | 预测单据ID |
