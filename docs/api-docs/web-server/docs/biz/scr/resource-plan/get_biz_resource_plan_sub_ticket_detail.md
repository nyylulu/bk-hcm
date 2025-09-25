### 描述

- 该接口提供版本：v1.8.5.6+。
- 该接口所需权限：业务访问。
- 该接口功能描述：获取资源预测申请子单详情。

### URL

GET /api/v1/woa/bizs/{bk_biz_id}/plans/resources/sub_tickets/{sub_ticket_id}

### 输入参数

无

### 调用示例

无

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "00000001",
    "base_info": {
      "type": "transfer",
      "type_name": "转移",
      "bk_biz_id": 123,
      "op_product_id": 1001,
      "plan_product_id": 1,
      "virtual_dept_id": 2,
      "submitted_at": "2019-07-29 11:57:20"
    },
    "status_info": {
      "status": "auditing",
      "status_name": "审批中",
      "stage": "crp_audit",
      "admin_audit_status": "skip",
      "crp_sn": "XQ000001",
      "crp_url": "http://crp/ticket/XQ000001",
      "message": "如果单据失败，这里会提供原因"
    },
    "demands": [
      {
        "demand_class": "CVM",
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
            "device_class": "标准型S5",
            "device_family": "标准型",
            "core_type": "大核心",
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
            "device_type": "SA3.2XLARGE16",
            "device_class": "标准型SA3",
            "device_family": "标准型",
            "core_type": "大核心",
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
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                        |
|---------|--------------|---------------------------|
| code    | int          | 错误编码。 0表示success，>0表示失败错误 |
| message | string       | 请求失败返回的错误信息               |
| data	   | object array | 响应数据                      |

#### data

| 参数名称        | 参数类型          | 描述           |
|-------------|---------------|--------------|
| id          | string	       | 资源预测需求子单ID   |
| base_info   | object	       | 资源预测需求子单基本信息 |
| status_info | object        | 资源预测需求子单状态信息 |
| demands     | object array	 | 资源预测需求列表     |

#### data.base_info

| 参数名称            | 参数类型   | 描述                                      |
|-----------------|--------|-----------------------------------------|
| type            | string | 单据类型（枚举值：add, adjust, cancel, transfer） |
| type_name       | string | 单据类型名称                                  |
| bk_biz_id       | int    | CC业务ID                                  |
| op_product_id   | int    | 运营产品ID                                  |
| plan_product_id | int    | 规划产品ID                                  |
| virtual_dept_id | int    | 虚拟部门ID                                  |
| submitted_at    | string | 提单时间                                    |

#### data.status_info

| 参数名称               | 参数类型   | 描述                                                        |
|--------------------|--------|-----------------------------------------------------------|
| status             | string | 单据状态（枚举值：init, auditing, rejected, failed, done, invalid） |
| status_name        | string | 单据状态名称（枚举值：待审批, 审批中, 审批拒绝, 失败, 成功, 已失效）                   |
| stage              | string | 单据审批阶段（枚举值：init（待审批）、admin_audit（部门审批）、crp_audit（公司审批））   |
| admin_audit_status | string | 管理员审批结果（枚举值：auditing、rejected、done、skip（自动通过））            |
| crp_sn             | string | CRP系统需求单号                                                 |
| crp_url            | string | CRP系统需求单链接                                                |
| message            | string | 单据状态失败信息                                                  |

#### data.demands[i]

| 参数名称          | 参数类型   | 描述                              |
|---------------|--------|---------------------------------|
| demand_class  | string | 预测的需求类型（枚举值：CVM, CA）            |
| original_info | object | 调整前需求信息（若为追加单，该参数内结构不变，值均为null） |
| updated_info  | object | 调整后需求信息（若为取消单，该参数内结构不变，值均为null） |

#### demands[i].original_info & demands[i].updated_info

| 参数名称             | 参数类型         | 描述                                                |
|------------------|--------------|---------------------------------------------------|
| obs_project      | string       | 项目类型                                              |
| expect_time      | string       | 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01                 |
| region_id        | string       | 地区/城市ID                                           |
| zone_id          | string       | 可用区ID                                             |
| demand_res_types | string array | 预测资源类型列表(枚举值：CVM、CBS)，需求包含CVM时，传递CVM，包含CBS时，传递CBS |
| cvm              | object       | 申请的CVM信息                                          |
| cbs              | object       | 申请的CBS信息                                          |

#### demands[i].original_info.cvm & demands[i].updated_info.cvm

| 参数名称          | 参数类型   | 描述                 |
|---------------|--------|--------------------|
| res_mode      | string | 资源模式(枚举值：按机型、按机型族) |
| device_type   | string | 机型规格               |
| device_class  | string | 机型分类               |
| device_family | string | 机型族                |
| core_type     | string | 核心类型(枚举值：大核心、小核心)  |
| os            | string | 实例数量，用string表示小数   |
| cpu_core      | int    | CPU核心数，单位：核        |
| memory        | int    | 内存大小，单位：GB         |

#### demands[i].original_info.cbs & demands[i].updated_info.cbs

| 参数名称           | 参数类型   | 描述                                                |
|----------------|--------|---------------------------------------------------|
| disk_type      | string | 云盘类型(枚举值：CLOUD_PREMIUM(高性能云硬盘)、CLOUD_SSD(SSD云硬盘)) |
| disk_type_name | string | 云盘类型名称                                            |
| disk_io        | int    | 磁盘IO吞吐需求，无特殊要求填写15；高性能云盘上限150，SSD云硬盘上限260         |
| disk_size      | int    | 云盘大小，单位：GB                                        |
