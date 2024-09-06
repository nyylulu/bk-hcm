### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：资源回收子任务步骤详情查询。

### URL

POST /api/v1/woa/task/find/recycle/step

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述            |
|-----------|----------|------|----------------|
| order_id  | int      | 是   | 资源回收单据ID   |
| task_id   | string   | 是	  | 资源回收子任务ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id":1001,
  "task_id":"1001-1"
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
        "step_id":2,
        "step_name":"checkUwork",
        "desciption":"检查是否有Uwork故障单据",
        "ip":"10.0.0.1",
        "retry_time": 5,
        "status":"FAILED",
        "message":"uwork have tickets, ticket number: `1`",
        "log":"",
        "create_at":"2022-01-02T15:04:05Z07:00",
        "update_at":"2022-01-02T15:04:05Z07:00"
      },
      {
        "step_id":1,
        "step_name":"preCheck",
        "desciption":"检查CC模块和负责人",
        "ip":"10.0.0.1",
        "retry_time": 0,
        "status":"SUCCESS",
        "message":"",
        "log":"",
        "create_at":"2022-01-02T15:04:05Z07:00",
        "update_at":"2022-01-02T15:04:05Z07:00"
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
| info    | object array | 回收子任务详情信息         |

#### data.info

| 参数名称     | 参数类型  | 描述               |
|-------------|---------|--------------------|
| step_id	  | int     | 资源回收步骤ID       |
| step_name	  | string  | 资源回收步骤名       |
| desciption  | string  | 资源回收步骤描述      |
| ip          | string  | 资源回收子任务主机IP  |
| retry_time  | int     | 重试次数             |
| status	  | string  | 资源回收步骤状态，"RUNNING": 执行中, "FAILED": 执行失败, "SUCCESS": 执行成功, "INIT": 待执行 |
| message	  | string  | 子任务结果详情         |
| log	      | string  | 步骤执行日志信息       |
| create_at	  | string  | 资源回收单据创建时间    |
| update_at	  | string  | 资源回收单据最后更新时间 |
