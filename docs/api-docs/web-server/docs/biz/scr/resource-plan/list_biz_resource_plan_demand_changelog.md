### 描述

- 该接口提供版本：v1.7.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询资源预测需求单的变更历史。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/plans/demands/change_logs/list

### 输入参数

| 参数名称          | 参数类型   | 必选 | 描述       |
|---------------|--------|----|----------|
| crp_demand_id | int    | 是  | 资源预测需求ID |
| page          | object | 是  | 分页设置     |

#### page

| 参数名称  | 参数类型 | 必选 | 描述                                                                                                                                        |
|-------|------|----|-------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start | int  | 否  | 记录开始位置，start 起始值为0                                                                                                                        |
| limit | int  | 否  | 每页限制条数，最大500，不能为0                                                                                                                         |

### 调用示例

```json
{
  "crp_demand_id": 387330,
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "details": [
      {
        "crp_demand_id": 387330,
        "expect_time": "2024-10-21",
        "bg_name": "IEG互动娱乐事业群",
        "dept_name": "IEG技术运营部",
        "plan_product_name": "移动终端游戏",
        "op_product_name": "运营产品",
        "obs_project": "常规项目",
        "region_name": "广州",
        "zone_name": "广州三区",
        "demand_week": "UNPLAN_9_13W",
        "res_pool_type": 0,
        "device_class": "标准型S5",
        "device_type": "S5.2XLARGE16",
        "change_cvm_amount": 0.125000,
        "after_cvm_amount": 0.125000,
        "change_core_amount": 1,
        "after_core_amount": 1,
        "change_ram_amount": 2,
        "after_ram_amount": 2,
        "disk_type": null,
        "disk_io": 0,
        "changed_disk_amount": 1,
        "after_disk_amount": 1,
        "demand_source": "追加需求订单",
        "crp_sn": "XQ202408221500512986",
        "create_time": null,
        "remark": "由 UNPLAN_9_13W 自动变为 UNPLAN_9_13W\n",
        "res_pool": "自研池"
      }
    ]
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

| 参数名称    | 参数类型         | 描述                                       |
|---------|--------------|------------------------------------------|
| count   | int          | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回 |
| details | object array | 查询返回的数据，仅在 count 查询参数设置为 false 时返回       |

#### data.details[n]

| 参数名称                | 参数类型   | 描述         |
|---------------------|--------|------------|
| crp_demand_id       | int    | CRP需求ID    |
| expect_time         | string | 期望交付时间     |
| bg_name             | string | 事业群        |
| dept_name           | string | 部门         |
| plan_product_name   | string | 规划产品       |
| op_product_name     | string | 运营产品       |
| obs_project         | string | 项目类型       |
| region_name         | string | 地区/城市名称    |
| zone_name           | string | 可用区        |
| demand_week         | string | 13周需求类型    |
| res_pool_type       | int    | 资源池类型ID    |
| res_pool            | string | 资源池        |
| device_class        | string | 机型类型       |
| device_type         | string | 机型规格       |
| change_cvm_amount   | int    | 实例数变更值     |
| after_cvm_amount    | int    | 实例数当前值     |
| change_core_amount  | int    | CPU核数变更值   |
| after_core_amount   | int    | CPU核数当前值   |
| change_ram_amount   | int    | 内存变更值（G）   |
| after_ram_amount    | int    | 内存当前值（G）   |
| changed_disk_amount | int    | 磁盘数变更值（G）  |
| after_disk_amount   | int    | 磁盘数当前值（G）  |
| disk_type           | string | 云盘类型       |
| disk_io             | int    | 磁盘IO（MB/s） |
| demand_source       | string | 变更类型       |
| crp_sn              | string | 变更单号       |
| create_time         | string | 变更时间       |
| remark              | string | 备注         |
