### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-资源预测。
- 该接口功能描述：管理员视角，查询资源预测需求单的变更历史。

### URL

POST /api/v1/woa/plans/demands/change_logs/list

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述     |
|-----------|--------|----|--------|
| demand_id | string | 是  | 预测需求ID |
| page      | object | 是  | 分页设置   |

#### page

| 参数名称  | 参数类型 | 必选 | 描述                                                                                                                                        |
|-------|------|----|-------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| start | int  | 否  | 记录开始位置，start 起始值为0                                                                                                                        |
| limit | int  | 否  | 每页限制条数，最大500，不能为0                                                                                                                         |

### 调用示例

```json
{
  "demand_id": "0000001z",
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
        "id": "00000011",
        "demand_id": "0000001z",
        "expect_time": "2024-10-21",
        "obs_project": "常规项目",
        "region_name": "广州",
        "zone_name": "广州三区",
        "device_type": "S5.2XLARGE16",
        "change_cvm_amount": "0.125000",
        "change_core_amount": 1,
        "change_ram_amount": 2,
        "changed_disk_amount": 1,
        "demand_source": "追加",
        "ticket_id": "00000022",
        "crp_sn": "XQ202408221500512986",
        "suborder_id": "",
        "create_time": "2024-09-01T12:00:00Z",
        "remark": "创建资源预测需求\n"
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

| 参数名称                | 参数类型   | 描述                                |
|---------------------|--------|-----------------------------------|
| id                  | string | 变更记录ID                            |
| demand_id           | string | 预测需求ID                            |
| expect_time         | string | 期望交付时间，格式为YYYY-MM-DD，例如2024-01-01 |
| obs_project         | string | 项目类型                              |
| region_name         | string | 地区/城市名称                           |
| zone_name           | string | 可用区                               |
| device_type         | string | 机型规格                              |
| change_cvm_amount   | string | 实例数变更值，可能为正或负                     |
| change_core_amount  | int    | CPU核数变更值，可能为正或负                   |
| change_ram_amount   | int    | 内存变更值（G），可能为正或负                   |
| changed_disk_amount | int    | 磁盘数变更值（G），可能为正或负                  |
| demand_source       | string | 变更类型，枚举值：追加、调整、删除、消耗              |
| ticket_id           | string | 变更需求的HCM订单号                       |
| crp_sn              | string | 变更需求的CRP订单号                       |
| suborder_id         | string | 主机申领的子订单号，在消耗记录用到                 |
| create_time         | string | 变更时间                              |
| remark              | string | 备注                                |
