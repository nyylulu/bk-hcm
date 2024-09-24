### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：业务-资源预测操作。
- 该接口功能描述：批量调整资源预测需求。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/demands/adjust

| 参数名称    | 参数类型         | 必选 | 描述   |
|---------|--------------|----|------|
| adjusts | object array | 是  | 调整列表 |

#### adjusts[i]

| 参数名称          | 参数类型   | 必选 | 描述                                                          |
|---------------|--------|----|-------------------------------------------------------------|
| crp_demand_id | int    | 是  | CRP需求ID                                                     |
| adjust_type   | string | 是  | 调整类型，枚举值：update（常规修改）、delay（加急延期）                           |
| demand_source | string | 否  | 需求分类/变更原因，adjust_type为update时必填                             |
| original_info | object | 否  | 调整前需求信息，adjust_type为update时必填                               |
| updated_info  | object | 否  | 调整后需求信息，adjust_type为update时必填                               |
| expect_time   | string | 否  | 修改后的期望交付时间，，adjust_type为delay时必填，格式为YYYY-MM-DD，例如2024-01-01 |
| delay_os      | int    | 否  | 延期OS数，adjust_type为delay时必填                                  |

#### adjusts[i].original_info & adjusts[i].updated_info

| 参数名称             | 参数类型         | 必选 | 描述                                                |
|------------------|--------------|----|---------------------------------------------------|
| obs_project      | string       | 是  | OBS项目类型                                           |
| expect_time      | string       | 是  | 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01                 |
| region_id        | string       | 是  | 地区/城市ID                                           |
| zone_id          | string       | 否  | 可用区ID                                             |
| demand_res_types | string array | 是  | 预测资源类型列表(枚举值：CVM、CBS)，需求包含CVM时，传递CVM，包含CBS时，传递CBS |
| cvm              | object       | 否  | 申请的CVM信息                                          |
| cbs              | object       | 否  | 申请的CBS信息                                          |

#### adjusts[i].original_info.cvm & adjusts[i].updated_info.cvm

| 参数名称        | 参数类型   | 必选 | 描述                 |
|-------------|--------|----|--------------------|
| res_mode    | string | 是  | 资源模式(枚举值：按机型、按机型族) |
| device_type | string | 是  | 机型规格               |
| os          | int    | 是  | OS数，单位：台           |
| cpu_core    | int    | 是  | CPU核心数，单位：核        |
| memory      | int    | 是  | 内存大小，单位：GB         |

#### adjusts[i].original_info.cbs & adjusts[i].updated_info.cbs

| 参数名称      | 参数类型   | 必选 | 描述                                                |
|-----------|--------|----|---------------------------------------------------|
| disk_type | string | 是  | 云盘类型(枚举值：CLOUD_PREMIUM(高性能云硬盘)、CLOUD_SSD(SSD云硬盘)) |
| disk_io   | int    | 是  | 磁盘IO吞吐需求，无特殊要求填写15；高性能云盘上限150，SSD云硬盘上限260         |
| disk_size | int    | 是  | 云盘大小，单位：GB                                        |

### 调用示例

```json
{
  "adjusts": [
    {
      "crp_demand_id": 387330,
      "adjust_type": "update",
      "demand_source": "指标变化",
      "original_info": {
        "obs_project": "常规项目",
        "expect_time": "2024-11-12",
        "region_id": "ap-shanghai",
        "zone_id": "ap-shanghai-2",
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
      },
      "updated_info": {
        "obs_project": "常规项目",
        "expect_time": "2024-11-12",
        "region_id": "ap-shanghai",
        "zone_id": "ap-shanghai-2",
        "demand_res_types": [
          "CVM",
          "CBS"
        ],
        "cvm": {
          "res_mode": "按机型",
          "device_type": "S5.2XLARGE16",
          "os": 1234,
          "cpu_core": 1234,
          "memory": 1234
        },
        "cbs": {
          "disk_type": "CLOUD_PREMIUM",
          "disk_io": 1234,
          "disk_size": 10245
        }
      }
    },
    {
      "crp_demand_id": 387330,
      "adjust_type": "delay",
      "expect_time": "2025-01-01",
      "delay_os": 10
    }
  ]
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
