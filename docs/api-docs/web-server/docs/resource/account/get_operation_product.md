### 描述

- 该接口提供版本：v1.6.0.0+。
- 该接口所需权限：
- 该接口功能描述：获取运营产品详情，返回运营产品详情（内部版）。

### URL

POST /api/v1/account/operation_products/{op_product_id}

## 请求参数
| 参数名称          | 参数类型 | 必选 | 描述     |
|---------------|------|----|--------|
| op_product_id | int  | 是  | 运营产品id |


### 响应数据
```
{
	"code": 0,
    "message": "",
    "data": {
		  "op_product_id": 5,						
		  "op_product_name": "xxx",					
		  "op_product_managers": "xxx,xxx,xxx",		
		  "op_product_bak_managers": "xxx,xxx,xxx"	
		  "plan_product_id": 55,					
		  "plan_product_name": "xxx"				
		  "bg_id": 1,							
		  "bg_name": "xxx",						
		  "bg_short_name": "xxx",					
		  "dept_id": 1,							
		  "dept_name": "xxx"						
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

| 参数名称                    | 参数类型   | 描述        |
|-------------------------|--------|-----------|
| op_product_id           | int    | 运营产品id    |
| op_product_name         | string | 运营产品Name  |
| op_product_managers     | string | 运营产品负责人   |
| op_product_bak_managers | string | 运营产品备份负责人 |
| plan_product_id         | int    | 规划产品id    |
| plan_product_name       | string | 规划产品Name  |
| bg_id                   | int    | bg id     |
| bg_name                 | string | bg名称      |
| bg_short_name           | string | 业务简称      |
| dept_id                 | int    | 部门id      |
| dept_name               | string | 部门名称      |
