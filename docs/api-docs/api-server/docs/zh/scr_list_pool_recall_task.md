### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询资源池下架任务。

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述        |
|--------------|--------------|------|------------|
| id           | int array    | 否   | 下架任务ID列表，数量最大20 |
| phase        | string array | 否   | 任务执行阶段列表 |
| start        | string	      | 否   | 单据创建时间过滤条件起点日期，格式如"2022-05-01" |
| end          | string	      | 否   | 单据创建时间过滤条件终点日期，格式如"2022-05-01" |
| page         | object	      | 是   | 分页信息     |

#### page

| 参数名称      | 参数类型 | 必选 | 描述                            |
|--------------|--------|-----|---------------------------------|
| start        | int    | 否  | 记录开始位置，start 起始值为0       |
| limit        | int    | 是  | 每页限制条数，最大200              |
| enable_count | bool   | 是  | 本次请求是否为获取数量还是详情的标记  |

注意：

- enable_count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

- 默认按id升序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "id":[
    1
  ],
  "phase":[
    "SUCCESS"
  ],
  "start":"2022-11-11",
  "end":"2022-11-11",
  "page":{
    "start":0,
    "limit":10
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
            "id":1,
            "spec":{
              "selector":[
                {
                  "key":"region",
                  "op":"equal",
                  "value":"南京"
                },
                {
                  "key":"zone",
                  "op":"equal",
                  "value":"南京-吉山"
                },
                {
                  "key":"resource_type",
                  "op":"equal",
                  "value":"IDCPM"
                },
                {
                  "key":"device_type",
                  "op":"equal",
                  "value":"CG2-10G"
                },
                {
                  "key":"ip",
                  "op":"in",
                  "value":[
                    "10.0.0.1",
                    "10.0.0.2"
                  ]
                }
              ],
              "replicas":2
            },
            "status":{
              "phase":"RUNNING",
              "message":"",
              "total_num":2,
              "success_num":1,
              "pending_num":1,
              "failed_num":0
            },
            "create_at":"2022-11-14T01:57:41.159Z",
            "update_at":"2022-11-14T01:57:41.159Z"
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
| info    | object array | 下架任务信息列表           |

#### info 字段说明：

| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
| id | int | 下架任务ID |
| spec | object | 下架任务明细 |
| status | object | 下架任务状态详情 |
| create_at | timestamp | 记录创建时间 |
| update_at | timestamp | 记录最后更新时间 |

#### spec 字段说明：
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
| selector | object array | 下架设备标签筛选条件 |
| replicas | int | 下架设备数量 |

#### spec.selector 字段说明：
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
| key | string | 标签键 |
| op | string | 操作符。"equal": 相等比较，"in": 匹配记录字段值是否在指定集合中 |
| value | interface{} | 标签值 |

#### status 字段说明：
| 参数名称     | 参数类型   | 描述                           |
| ------------ | ---------- | ------------------------------ |
| phase | string | 下架任务执行阶段。INIT：待执行，RUNNING：执行中，PAUSED：已暂停，SUCCESS：执行成功，FAILED：执行失败 |
| message | string | 下架任务信息 |
| total_num | int | 总下架设备数 |
| success_num | int | 成功下架设备数 |
| pending_num | int | 待下架设备数 |
| failed_num | int | 下架失败设备数 |

**注意：**
- 如果本次请求是查询详细信息那么count为0，如果查询的是数量，那么info为空。
