### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源初始化记录查询。

### URL

POST /api/v1/woa/task/find/apply/record/init

### 输入参数

| 参数名称     | 参数类型  | 必选 | 描述         |
|-------------|---------|------|-------------|
| suborder_id | string  | 是   | 资源申请子单号 |
| filter      | object  | 是   | 查询过滤条件   |
| page        | object  | 是   | 分页设置      |

#### filter

| 参数名称   | 参数类型      | 必选 | 描述                                                              |
|-----------|-------------|------|------------------------------------------------------------------|
| condition | enum string | 是   | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rule      | array       | 是   | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。 |

#### rule[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型      | 必选 | 描述                                          |
|----------|-------------|----|------------------------------------------------|
| field    | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| operator | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value    | 可变类型     | 是  | 查询条件Value值                                  |

##### rule 表达式说明：

##### 1. 操作符

| 操作符            | 描述                                        | 操作符的value支持的数据类型                                 |
|------------------|---------------------------------------------|---------------------------------------------------------|
| equal            | 等于。不能为空字符串                           | boolean, numeric, string                                |
| not_equal        | 不等。不能为空字符串                           | boolean, numeric, string                                |
| greater          | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| greater_or_equal | 大于等于                                     | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| less             | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| less_or_equal    | 小于等于                                     | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in               | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string             |
| not_in           | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string            |
| contains         | 模糊查询，区分大小写                           | string                                                   |

#### page

| 参数名称 | 参数类型 | 必选 | 描述                            |
|---------|--------|-----|---------------------------------|
| start   | int    | 否  | 记录开始位置，start 起始值为0       |
| limit   | int    | 是  | 每页限制条数，最大200              |
| sort    | string | 否  | 排序字段，返回数据将按该字段进行排序  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "suborder_id": "1001-1",
  "page":{
    "start":0,
    "limit":100,
    "sort":"ip"
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
        "suborder_id":"1001-1",
        "ip":"10.0.0.1",
        "task_id":"5684804",
        "task_link":"http://scr.ied.com/#/sojob/list/5684804",
        "status":0,
        "message":"success",
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
| permission | object       | 权限信息             |
| request_id | string       | 请求链ID             |
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 资源初始化记录信息列表      |

#### data.info

| 参数名称       | 参数类型    | 描述          |
|---------------|-----------|---------------|
| suborder_id   | string    | 资源申请子单号  |
| ip	        | string	| 资源IP        |
| task_id	    | string	| 初始化任务ID   |
| task_link	    | string	| 初始化任务详情链接   |
| status	    | int	    | 初始化记录状态码，0: 成功, 1: 执行中, 其他: 失败 |
| message	    | string	| 初始化记录状态信息   |
| start_at	    | timestamp	| 生产记录开始时间    |
| end_at	    | timestamp	| 生产记录结束时间    |
