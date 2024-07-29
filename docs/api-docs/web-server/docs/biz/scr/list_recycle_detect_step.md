### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源回收预检任务步骤查询。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/findmany/recycle/detect/step

### 输入参数

| 参数名称      | 参数类型           | 必选 | 描述                       |
|--------------|------------------|------|---------------------------|
| order_id	   | int array	      | 否   | 资源回收单号列表，数量最大20   |
| suborder_id  | string	array     | 否   | 资源回收子单号列表，数量最大20 |
| ip           | string array     | 否   | 设备内网IP列表，数量最大500   |
| step_name    | string	array     | 否   | 预检步骤列表，数量最大20      |
| status       | string	array     | 否   | 预检状态列表，数量最大20      |
| bk_username  | string	array     | 否   | 提单人列表，数量最大20        |
| start        | string	          | 否   | 单据创建时间过滤条件起点日期，格式如"2022-05-01" |
| end          | string	          | 否   | 单据创建时间过滤条件终点日期，格式如"2022-05-01" |
| page         | object	          | 是   | 分页信息                    |

#### page

| 参数名称      | 参数类型 | 必选 | 描述                            |
|--------------|--------|-----|---------------------------------|
| start        | int    | 否  | 记录开始位置，start 起始值为0       |
| limit        | int    | 是  | 每页限制条数，最大100              |
| enable_count | bool   | 是  | 本次请求是否为获取数量还是详情的标记  |

说明：

- enable_count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

- 默认按create_at降序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id":[1],
  "page":{
    "start":0,
    "limit": 20,
    "enable_count":false
  }
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
    "count":1,
    "info":[
      {
        "order_id":1,
        "suborder_id":"1-1",
        "ip":"10.0.0.1",
        "step_id":1,
        "step_name":"PRECHECK",
        "step_desc":"检查CC模块和负责人",
        "retry_time":0,
        "status":"DONE",
        "message":"success",
        "log":"",
        "start_at":"2022-04-25T15:04:05.004Z",
        "end_at":"2022-04-25T15:04:05.004Z"
        "create_at":"2022-04-25T15:04:05.004Z",
        "update_at":"2022-04-25T15:04:05.004Z"
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
| info    | object array | 资源回收单据信息列表     |

#### data.info

| 参数名称     | 参数类型    | 描述                |
|-------------|-----------|---------------------|
| order_id	  | int	      | 资源回收单号          |
| suborder_id | string	  | 资源回收子单号        |
| ip	      | int	      | 主机IP               |
| step_id	  | int	      | 预检任务步骤ID        |
| step_name	  | string	  | 预检任务步骤名        |
| step_desc	  | string	  | 预检任务步骤描述      |
| retry_time  | int	      | 重试次数             |
| status	  | string	  | 预检任务状态          |
| message	  | string	  | 预检任务结果详情       |
| log	      | string	  | 预检任务步骤执行日志信息 |
| start_at	  | timestamp | 预检任务步骤开始时间    |
| end_at	  | timestamp | 预检任务步骤结束时间    |
| create_at	  | timestamp | 预检任务步骤记录创建时间    |
| update_at	  | timestamp | 预检任务步骤记录最后更新时间 |

注意：

- 如果本次请求是查询详细信息那么count为0，如果查询的是数量，那么info为空。
