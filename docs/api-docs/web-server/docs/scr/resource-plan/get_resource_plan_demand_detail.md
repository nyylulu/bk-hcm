### 描述

- 该接口提供版本：1.7.1+
- 该接口所需权限：平台-资源预测。
- 该接口功能描述：管理员视角，获取CRP平台的资源预测需求单详情。

### URL

GET /api/v1/woa/plans/demands/{id}

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
    "crp_demand_id": "387330",
    "year_month_week": "2024年10月4周",
    "expect_start_date": "2024-10-21",
    "expect_end_date": "2024-10-27",
    "expect_time": "2024-10-21",
    "bg_id": 4,
    "bg_name": "IEG互动娱乐事业群",
    "dept_id": 1041,
    "dept_name": "IEG技术运营部",
    "plan_product_id": 34,
    "plan_product_name": "规划产品",
    "op_product_id": 41,
    "op_product_name": "运营产品",
    "obs_project": "常规项目",
    "region_id": 73,
    "region_name": "广州",
    "zone_id": 100003,
    "zone_name": "广州三区",
    "plan_type": "计划内",
    "plan_advance_week": 9,
    "expedited_postponed": "无变化",
    "core_type_id": 1,
    "core_type": "小核心",
    "device_family": "标准型",
    "device_class": "标准型S5",
    "device_type": "S5.2XLARGE16",
    "disk_io": 150,
    "disk_type_id": 606,
    "disk_type": "高性能云硬盘",
    "demand_week": "UNPLAN_9_13W",
    "res_pool_type": 0,
    "res_pool": "自研池",
    "res_mode": "按机型",
    "generation_type": "采购"
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

| 参数名称                | 参数类型    | 描述            |
|---------------------|---------|---------------|
| crp_demand_id       | string	 | CRP需求ID       |
| year_month_week     | string  | 需求年月周         |
| expect_start_date   | string  | 期望最早到货时间      |
| expect_end_date     | string  | 期望最晚到货时间      |
| expect_time         | string  | 期望到货时间        |
| bg_id               | int     | 事业群ID         |
| bg_name             | string  | 事业群           |
| dept_id             | int     | 部门ID          |
| dept_name           | string  | 部门            |
| plan_product_id     | int     | 规划产品ID        |
| plan_product_name   | string  | 规划产品          |
| op_product_id       | int     | 运营产品ID        |
| op_product_name     | string  | 运营产品          |
| obs_project         | string  | 项目类型          |
| region_id           | int     | 地区/城市ID       |
| region_name         | string  | 地区/城市         |
| zone_id             | int     | 可用区ID         |
| zone_name           | string  | 期望可用区         |
| plan_type           | string  | 计划类型          |
| plan_advance_week   | int     | 计划提前周         |
| expedited_postponed | string  | 加急延期类型        |
| core_type_id        | int     | 核心类型ID        |
| core_type           | string  | 核心类型          |
| device_family       | string  | 机型族           |
| device_class        | string  | 机型类型          |
| device_type         | string  | 机型规格          |
| disk_io             | int     | 单实例磁盘IO(MB/s) |
| disk_type           | string  | 云盘类型          |
| demand_week         | string  | 13周需求类型       |
| res_pool_type       | int     | 资源池类型         |
| res_pool            | string  | 资源池           |
| res_mode            | string  | 资源模式          |
| generation_type     | string  | 机型代次          |
