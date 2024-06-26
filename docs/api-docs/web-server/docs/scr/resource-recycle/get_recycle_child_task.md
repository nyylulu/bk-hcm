### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源回收单据子任务详情查询。

### URL

POST /api/v1/woa/task/find/recycle/task

### 输入参数

| 参数名称   | 参数类型       | 必选 | 描述          |
|-----------|--------------|------|--------------|
| order_id  | int          | 是   | 资源回收单号   |
| status    | string array | 否	  | 资源回收单据状态，"RUNNING": 执行中, "FAILED": 失败, "SUCCESS": 成功, "INIT": 未执行 |
| last_step	| string array | 否	  | 最后执行的步骤 |
| page	    | object       | 是	  | 分页信息      |

#### page

| 参数名称      | 参数类型 | 必选 | 描述                            |
|--------------|--------|-----|---------------------------------|
| start        | int    | 否  | 记录开始位置，start 起始值为0       |
| limit        | int    | 是  | 每页限制条数，最大200              |
| sort	       | string	| 否  | 排序字段                          |
| enable_count | bool   | 是  | 本次请求是否为获取数量还是详情的标记  |

**注意：**

- enable_count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

- 默认按task_id降序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id": 1001,
  "page":{
    "start":0,
    "limit":100,
    "sort":"task_id",
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
  "permission":null,
  "request_id":"f5a6331d4bc2433587a63390c76ba7bf",
  "data":{
    "count":1,
    "info":[
      {
        "order_id":1001,
        "task_id":"1001-1",
        "ip":"10.0.0.1",
        "bk_username":"xxx",
        "status":"FAILED",
        "progress":"7/10",
        "last_step":"空闲检查",
        "message":"Exec TimeOut:"
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
| permission | object       | 权限信息             |
| request_id | string       | 请求链ID             |
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 回收单据子任务详情信息      |

#### data.info

| 参数名称     | 参数类型  | 描述               |
|-------------|---------|--------------------|
| order_id    | int     | 资源回收单据ID       |
| task_id     | string	| 资源回收子任务ID     |
| ip          | string  | 资源回收子任务主机IP  |
| bk_username | string	| 资源回收任务提单人    |
| create_at	  | string	| 资源回收单据创建时间   |
| update_at	  | string	| 资源回收单据最后更新时间 |
| status	  | string	| 资源回收单据状态，"RUNNING": 有步骤执行中, "FAILED": 有步骤执行失败, "SUCCESS": 全部步骤执行成功, "INIT": 所有步骤待执行 |
| progress	  | string	| 回收任务进度，格式为: 已成功的步骤数/总步骤数 |
| last_step	  | string	| 最后执行的步骤 |
| message	  | string	| 子任务结果详情 |

**注意：**

- 如果本次请求是查询详细信息那么count为0，如果查询的是数量，那么info为空。
