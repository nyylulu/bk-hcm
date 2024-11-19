### 描述

- 该接口提供版本：v9.9.9+
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
    "demand_id": "0000001z",
    "expect_time": "2024-10-21",
    "bk_biz_id": 111,
    "bk_biz_name": "业务",
    "dept_id": 1041,
    "dept_name": "IEG技术运营部",
    "plan_product_id": 34,
    "plan_product_name": "规划产品",
    "op_product_id": 41,
    "op_product_name": "运营产品",
    "obs_project": "常规项目",
    "area_id": "south",
    "area_name": "华南地区",
    "region_id": "guangzhou",
    "region_name": "广州",
    "zone_id": "guangzhou-3",
    "zone_name": "广州三区",
    "plan_type": "预测内",
    "core_type": "小核心",
    "device_family": "标准型",
    "device_class": "标准型S5",
    "device_type": "S5.2XLARGE16",
    "os": "0.125000",
    "memory": 2,
    "cpu_core": 1,
    "disk_size": 1,
    "disk_io": 150,
    "disk_type": "CLOUD_PREMIUM",
    "disk_type_name": "高性能云硬盘"
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

| 参数名称              | 参数类型    | 描述            |
|-------------------|---------|---------------|
| demand_id         | string	 | 预测需求ID        |
| expect_time       | string  | 期望到货时间        |
| bk_biz_id         | int     | 业务ID          |
| bk_biz_name       | string  | 业务            |
| dept_id           | int     | 部门ID          |
| dept_name         | string  | 部门            |
| plan_product_id   | int     | 规划产品ID        |
| plan_product_name | string  | 规划产品          |
| op_product_id     | int     | 运营产品ID        |
| op_product_name   | string  | 运营产品          |
| obs_project       | string  | 项目类型          |
| area_id           | string  | 区域ID          |
| area_name         | string  | 区域名称          |
| region_id         | string  | 地区/城市ID       |
| region_name       | string  | 地区/城市         |
| zone_id           | string  | 可用区ID         |
| zone_name         | string  | 期望可用区         |
| plan_type         | string  | 计划类型          |
| core_type         | string  | 核心类型          |
| device_family     | string  | 机型族           |
| device_class      | string  | 机型类型          |
| device_type       | string  | 机型规格          |
| os                | string  | 实例数           |
| memory            | int     | 总内存（G）        |
| cpu_core          | int     | 总CPU（核）       |
| disk_size         | int     | 总云盘大小（G）      |
| disk_io           | int     | 单实例磁盘IO(MB/s) |
| disk_type         | string  | 云盘类型          |
| disk_type_name    | string  | 云盘类型中文名       |
