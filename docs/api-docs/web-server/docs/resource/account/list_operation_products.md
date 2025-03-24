### 描述

- 该接口提供版本：v1.6.0.0+。
- 该接口所需权限：
- 该接口功能描述：运营产品拉取（内部版）。

### URL

POST /api/v1/account/operation_products/list

## 请求参数
| 参数名称            | 参数类型      | 必选 | 描述                  |
|-----------------|-----------|----|---------------------|
| op_product_ids  | int array | 否  | 运营产品id列表,最多支持传入500个 |
| op_product_name | string    | 否  | 运营产品名称，支持模糊搜索       |
| dept_ids        | int array | 否  | 部门id列表,最多支持传入20个    |
| bg_ids          | int array | 否  | 事业群id列表,最多支持传入20个   |
| page            | object    | 是  | 分页设置                |


#### Page
| 参数名称   | 参数类型    | 必选 | 描述                                                                                                                                               |
|--------|---------|----|--------------------------------------------------------------------------------------------------------------------------------------------------|
| count  | bool    | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据 detail，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count |
| limit  | uint    | 是  | 每页限制条数，最大500，不能为0                                                                                                                                |
| start  | uint    | 否  | 记录开始位置，start 起始值为0                                                                                                                               |
| sort	  | string	 | 否	 | 排序字段，返回数据将按该字段进行排序                                                                                                                               |
| order	 | string	 | 否	 | 排序顺序（枚举值：ASC、DESC）                                                                                                                               |

```
{
	// json body
	"op_product_name": "运营产品名",
	"op_product_ids":[1,2,4]
}
```

### 响应数据
```
{
	"code": 0,
    "message": "",
    "data": [
        {
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
    ]
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | array  | 响应数据 |

#### data[i]

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
