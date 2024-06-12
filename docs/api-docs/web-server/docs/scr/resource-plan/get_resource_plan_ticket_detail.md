### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：获取资源预测申请单据详情。

### URL

GET /api/v1/woa/plan/resource/ticket/{id}

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
      "applicant": "abc",
      "bk_biz_id": 123,
      "bk_biz_name": "biz_test",
      "bk_product_id": 1001,
      "bk_product_name": "运营产品",
      "plan_product_id": 1,
      "plan_product_name": "规划产品",
      "virtual_dept_id": 2,
      "virtual_dept_name": "部门",
      "demand_class": "CVM",
      "remark": "这里是预测说明",
      "submitted_at": "2019-07-29 11:57:20"
    },
    "status_info": {
      "status": "AUDIT",
      "status_name": "审批中",
      "itsm_sn": "REQ000001",
      "itsm_url": "http://itsm/ticket/REQ000001",
      "crp_sn": "XQ000001",
      "crp_url": "http://crp/ticket/XQ000001"
    },
    "demands": [
      {
        "obs_project": "常规项目",
        "expect_time": "2024-11-12",
        "area_id": "2",
        "area_name": "华东地区",
        "region_id": "ap-shanghai",
        "region_name": "上海",
        "zone_id": "ap-shanghai-2",
        "zone_name": "上海二区",
        "res_mode": "按机型",
        "demand_source": "指标变化",
        "remark": "这里是需求备注",
        "cvm": {
          "res_mode": "按机型",
          "device_family": "标准型",
          "device_type": "S5.2XLARGE16",
          "device_class": "标准型S5",
          "cpu_core": 123,
          "memory": 123,
          "res_pool": "自研池",
          "core_type": "大核心"
        },
        "cbs": {
          "disk_type": "CLOUD_PREMIUM",
          "disk_type_name": "高性能云硬盘",
          "disk_io": 123,
          "disk_size": 1024
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

| 参数名称        | 参数类型          | 描述          |
|-------------|---------------|-------------|
| id          | string	       | 资源预测申请单号    |
| base_info   | object	       | 资源预测申请单基本信息 |
| status_info | object        | 资源预测申请单状态信息 |
| demands     | object array	 | 资源预测需求列表    |

#### data.base_info

| 参数名称              | 参数类型   | 描述      |
|-------------------|--------|---------|
| applicant         | string | 申请人     |
| bk_biz_id         | int	   | CC业务ID  |
| bk_biz_name       | string | CC业务名   |
| bk_product_id     | int    | 运营产品ID  |
| bk_product_name   | string | 运营产品名称  |
| plan_product_id   | int    | 规划产品ID  |
| plan_product_name | string | 规划产品名称  |
| virtual_dept_id   | int    | 虚拟部门ID  |
| virtual_dept_name | string | 虚拟部门名称  |
| demand_class      | string | 预测的需求类型 |
| remark            | string | 预测说明    |
| submitted_at      | string | 提单时间    |

#### data.status_info

| 参数名称        | 参数类型   | 描述                                               |
|-------------|--------|--------------------------------------------------|
| status      | string | 单据状态（枚举值：init, auditing, rejected, done, failed） |
| status_name | string | 单据状态名称                                           |
| itsm_sn     | string | ITSM流程单号                                         |
| itsm_url    | string | ITSM流程单链接                                        |
| crp_sn      | string | CRP系统需求单号                                        |
| crp_url     | string | CRP系统需求单链接                                       |

#### data.demands[i]

| 参数名称             | 参数类型   | 描述                                |
|------------------|--------|-----------------------------------|
| obs_project      | string | OBS项目类型                           |
| expect_time      | string | 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01 |
| demand_week      | string | 13周需求类型，由CRP系统定义                  |
| demand_week_name | string | 13周需求类型名称                         |
| area_id          | string | 区域ID                              |
| area_name        | string | 区域名称                              |
| region_id        | string | 地区/城市ID                           |
| region_name      | string | 地区/城市名称                           |
| zone_id          | string | 可用区ID                             |
| zone_name        | string | 可用区名称                             |
| demand_source    | string | 需求分类/变更原因                         |
| remark           | string | 需求备注                              |
| cvm              | object | 申请的CVM信息                          |
| cbs              | object | 申请的CBS信息                          |

#### data.demands[i].cvm

| 参数名称          | 参数类型   | 描述                 |
|---------------|--------|--------------------|
| res_mode      | string | 资源模式(枚举值：按机型、按机型族) |
| device_family | string | 机型族                |
| device_type   | string | 机型规格               |
| device_class  | string | 机型分类               |
| cpu_core      | int    | CPU核心数，单位：核        |
| memory        | int    | 内存大小，单位：GB         |
| res_pool      | string | 资源池(枚举值：自研池、公有池)   |
| core_type     | string | 核心类型(枚举值：大核心、小核心)  |

#### data.demands[i].cbs

| 参数名称           | 参数类型   | 描述                                                |
|----------------|--------|---------------------------------------------------|
| disk_type      | string | 云盘类型(枚举值：CLOUD_PREMIUM(高性能云硬盘)、CLOUD_SSD(SSD云硬盘)) |
| disk_type_name | string | 云盘类型名称                                            |
| disk_io        | int    | 磁盘IO吞吐需求，无特殊要求填写15；高性能云盘上限150，SSD云硬盘上限260         |
| disk_size      | int    | 云盘大小，单位：GB                                        |
