### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源申请单据详情查询。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/find/apply/detail

### 输入参数

| 参数名称      | 参数类型 | 必选 | 描述          |
|-------------|---------|------|--------------|
| suborder_id | string  | 是   | 资源申请子单号  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "suborder_id": "1001-1"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data":{
    "count":2,
    "info":[
      {
        "suborder_id":"1001-1",
        "step_id":1,
        "step_name":"下单",
        "status":0,
        "message":"success",
        "total_num":10,
        "success_num":10,
        "failed_num":0,
        "running_num":0,
        "start_at":"2022-01-02T15:04:05.004Z",
        "end_at":"2022-01-02T15:04:05.004Z"
      },
      {
        "suborder_id":"1001-1",
        "step_id":2,
        "step_name":"生产",
        "status":0,
        "message":"success",
        "total_num":10,
        "success_num":10,
        "failed_num":0,
        "running_num":0,
        "start_at":"2022-01-02T15:04:05.004Z",
        "end_at":"2022-01-02T15:04:05.004Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 资源单据步骤信息列表        |

#### data.info

| 参数名称      | 参数类型   | 描述         |
|-------------|-----------|--------------|
| suborder_id | string	  | 资源申请子单号 |
| step_id	  | int	      | 步骤ID        |
| step_name   | string	  | 步骤名        |
| status	  | int       | 步骤状态码，0: 成功, 1: 执行中, 其他: 失败 |
| message	  | string    | 步骤状态信息   |
| total_num	  | int	      | 资源需求总数   |
| success_num | int	      | 当前步骤成功的资源数
| start_at	  | timestamp | 步骤开始时间   |
| end_at	  | timestamp |	步骤结束时间   |
