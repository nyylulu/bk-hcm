### 描述

- 该接口提供版本：v1.8.5.0+。
- 该接口所需权限：业务访问。
- 该接口功能描述：分页查询指定空闲检查子单下的机器空闲检查任务执行信息

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/idle_check/result/{suborder_id}

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述            |
|-------------|--------|----|---------------|
| bk_biz_id   | int64  | 是  | 业务ID          |
| suborder_id | string | 是  | 待查询结果的空闲检查子单号 |
| page        | object | 是  | 分页设置          |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 否  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0（查看第start台～第start+limit台机器的空闲检查结果）                                                                                                |
| limit | uint32 | 是  | 每次查询结果的机器数限制，最大50台（因为每台机器空闲检查时都有10个步骤，此时响应结果长度为50，分别描述了50台机器的空闲检查结果，每一个元素中包含了这台机器空闲检查任务执行的整体情况以及长度为10的空闲检查步骤具体执行情况描述）                               |

### 调用示例

#### 获取空闲检查结果请求示例

```json
{
  "page": {
    "count": false,
    "start": 0,
    "limit": 50
  }
}
```

### 响应示例

#### 获取空闲检查结果返回示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "count": 0,
    "details": [
      {
        "detect_task": {
          "task_id": "xxx",
          "order_id": 123,
          "suborder_id": "xxx",
          "bk_asset_id": "xxx",
          "bk_host_id": 12345,
          "ip": "10.0.0.1",
          "bk_username": "admin",
          "status": "SUCCESS",
          "message": "xxx",
          "total_num": 5,
          "success_num": 5,
          "pending_num": 0,
          "failed_num": 0,
          "create_at": "2024-08-22T10:00:00Z",
          "update_at": "2024-08-22T10:05:00Z"
        },
        "detect_steps": [
          {
            "id": "xxx",
            "order_id": 123,
            "suborder_id": "xxx",
            "task_id": "xxx",
            "step_id": 1,
            "step_name": "xxx",
            "step_desc": "xxx",
            "bk_host_id": 12345,
            "bk_asset_id": "xxx",
            "ip": "10.0.0.1",
            "bk_username": "admin",
            "retry_time": 0,
            "status": "SUCCESS",
            "message": "xxx",
            "skip": 0,
            "log": "xxx",
            "start_at": "2024-08-22T10:00:00Z",
            "end_at": "2024-08-22T10:01:00Z",
            "create_at": "2024-08-22T10:00:00Z",
            "update_at": "2024-08-22T10:01:00Z"
          },
          {
            "id": "xxx",
            "order_id": 123,
            "suborder_id": "xxx",
            "task_id": "xxx",
            "step_id": 2,
            "step_name": "xxx",
            "step_desc": "xxx",
            "bk_host_id": 12345,
            "bk_asset_id": "xxx",
            "ip": "10.0.0.1",
            "bk_username": "admin",
            "retry_time": 0,
            "status": "SUCCESS",
            "message": "xxx",
            "skip": 0,
            "log": "xxx",
            "start_at": "2024-08-22T10:01:00Z",
            "end_at": "2024-08-22T10:02:00Z",
            "create_at": "2024-08-22T10:01:00Z",
            "update_at": "2024-08-22T10:02:00Z"
          }
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                         |
|---------|--------|----------------------------|
| result  | bool   | 请求成功与否。true:请求成功；false请求失败 |
| code    | int    | 错误编码。 0表示success，>0表示失败错误  |
| message | string | 请求失败返回的错误信息                |
| data    | object | 响应数据                       |

#### data

| 参数名称    | 参数类型         | 描述                                                       |
|---------|--------------|----------------------------------------------------------|
| count   | uint64       | 当前规则能匹配到的总记录条数                                           |
| details | object array | 本次查询的每台机器的整体检查结果及其对应的每个检查步骤的详细执行情况,长度为本次查询的机器数，小于等于limit |

#### data.details[i]

查询到的第i个执行空闲检查任务的机器对应的空闲检查任务整体执行信息，以及每个空闲检查步骤的执行详情
1台待空闲检查主机->1个detect_task->10个detect_step

| 参数名称         | 参数类型         | 描述                                                                |
|--------------|--------------|-------------------------------------------------------------------|
| detect_task  | object       | 某台执行空闲检查任务的机器的整体执行信息，比如执行成功的步骤数                                   |
| detect_steps | object array | 空闲检查步骤详情列表，长度为空闲检查步骤数（每个元素表示这台执行空闲检查任务的机器在某个步骤的详细信息，比如这个步骤是否执行成功） |

#### detect_task

| 字段名称        | 类型     | 描述                                           |
|-------------|--------|----------------------------------------------|
| task_id     | string | 任务ID                                         |
| order_id    | uint64 | 订单ID                                         |
| suborder_id | string | 子订单ID                                        |
| bk_asset_id | string | 设备固资号                                        |
| bk_host_id  | int64  | 主机ID                                         |
| ip          | string | IP地址                                         |
| bk_username | string | 操作用户                                         |
| status      | string | 任务状态（枚举值：INIT、RUNNING、PAUSED、SUCCESS、FAILED） |
| message     | string | 任务消息                                         |
| total_num   | uint   | 总步骤数                                         |
| success_num | uint   | 成功步骤数                                        |
| pending_num | uint   | 待处理步骤数                                       |
| failed_num  | uint   | 失败步骤数                                        |
| create_at	  | string | 空闲检查任务步骤记录创建时间                               |
| update_at	  | string | 空闲检查任务步骤记录最后更新时间                             |

#### detect_steps[i]

| 字段名称        | 类型     | 描述                                           |
|-------------|--------|----------------------------------------------|
| id          | string | 步骤ID                                         |
| order_id    | uint64 | 订单ID                                         |
| suborder_id | string | 子订单ID                                        |
| task_id     | string | 任务ID                                         |
| step_id     | int    | 步骤序号                                         |
| step_name   | string | 步骤名称                                         |
| step_desc   | string | 步骤描述                                         |
| bk_host_id  | int64  | 主机ID                                         |
| bk_asset_id | string | 设备固资号                                        |
| ip          | string | IP地址                                         |
| bk_username | string | 操作用户                                         |
| retry_time  | uint32 | 重试次数                                         |
| status      | string | 任务状态（枚举值：INIT、RUNNING、PAUSED、SUCCESS、FAILED） |
| message     | string | 步骤消息                                         |
| skip        | int    | 是否跳过（0:否 1:是）                                |
| log         | string | 日志信息                                         |
| start_at	   | string | 空闲检查任务步骤开始时间                                 |
| end_at	     | string | 空闲检查任务步骤结束时间                                 |
| create_at	  | string | 空闲检查任务步骤记录创建时间                               |
| update_at	  | string | 空闲检查任务步骤记录最后更新时间                             |