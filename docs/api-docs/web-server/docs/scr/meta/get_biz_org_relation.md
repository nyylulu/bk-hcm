### 描述

- 该接口提供版本：v1.5.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务组织关系。

### URL

GET /api/v1/woa/bizs/{bk_biz_id}/org/relation

### 输入参数

| 参数名称      | 参数类型 | 必选 | 描述   |
|-----------|------|----|------|
| bk_biz_id | int  | 是  | 业务ID |

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "bk_biz_id": 111,
    "bk_biz_name": "业务",
    "op_product_id": 222,
    "op_product_name": "运营产品",
    "plan_product_id": 333,
    "plan_product_name": "规划产品",
    "virtual_dept_id": 1041,
    "virtual_dept_name": "IEG技术运营部"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int    | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称              | 参数类型   | 描述     |
|-------------------|--------|--------|
| bk_biz_id         | int    | 业务ID   |
| bk_biz_name       | string | 业务名称   |
| op_product_id     | int    | 运营产品ID |
| op_product_name   | string | 运营产品名称 |
| plan_product_id   | int    | 规划产品ID |
| plan_product_name | string | 规划产品名称 |
| virtual_dept_id   | int    | 虚拟部门ID |
| virtual_dept_name | string | 虚拟部门名称 |
