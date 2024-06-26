### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：机房裁撤。
- 该接口功能描述：更新裁撤模块信息。

### URL

PUT /api/v1/woa/dissolve/recycled_module/update

### 输入参数

| 参数名称         | 参数类型    | 必选 | 描述                                 |
|--------------|---------|----|------------------------------------|
| id           | string	 | 是	 | 模块唯一标识                             |
| name         | string	 | 否	 | 模块名称                               |
| start_time   | string  | 否	 | 开始日期，格式为yyyy-mm-dd                 |
| end_time     | string  | 否	 | 结束日期，格式为yyyy-mm-dd                 |
| which_stages | int	    | 否	 | 裁撤模块信息分类，用于前端展示                    |
| recycle_type | int     | 否	 | 裁撤类型，recycle_type有两个值，0表示全裁，1表示部分裁 |


### 调用示例

```json
{
  "id": "1",
  "name": "深圳-锦绣-M11",
  "start_time": "2023-01-01",
  "end_time": "2023-05-31",
  "which_stages": 12,
  "recycle_type": 0
}
```

### 响应示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "permission":null,
  "request_id":"f5a6331d4bc2433587a63390c76ba7bf",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型      | 描述               |
|------------|-------------|--------------------|
| result     | bool        | 请求成功与否。true:请求成功；false请求失败 |
| code       | int         | 错误编码。 0表示success，>0表示失败错误  |
| message    | string      | 请求失败返回的错误信息 |
| permission | object      | 权限信息             |
| request_id | string      | 请求链ID             |
| data	     | object | 响应数据             |
