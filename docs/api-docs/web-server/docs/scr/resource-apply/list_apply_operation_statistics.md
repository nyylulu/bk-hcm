### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源申请运营统计查询。

### URL

POST /api/v1/woa/task/find/operation/apply/statistics

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述      |
|-----------|--------|-----|-----------|
| start	    | string | 是  | 时间范围起点日期，格式如"2022-05-01" |
| end	    | string | 否  | 时间范围终点日期，格式如"2022-05-01" |
| dimension	| string | 否  | 时间维度。"DAY":按天统计, "MONTH":按月统计, "YEAR": 按年统计 |
| filter    | object | 否  | 查询过滤条件 |

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

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "start":"2022-07-11",
  "end":"2022-07-11",
  "dimension":"DAY",
  "filter":{
    "condition":"AND",
    "rules":[
      {
        "field":"resource_type",
        "operator":"equal",
        "value":"QCLOUDCVM"
      }
    ]
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
        "date":"2022-07-11",
        "order_total": 100,
        "order_succ": 98,
        "order_succ_rate": 0.98,
        "os_total": 1000,
        "os_succ": 990,
        "os_succ_rate": 0.99
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
| info    | object array | 资源申请运营统计信息列表    |

#### data.info

| 参数名称         | 参数类型 | 描述            |
|-----------------|--------|-----------------|
| date	          | string | 日期。按天统计的格式如"2022-05-01"，按月统计的格式如"2022-05"，按年统计的格式如"2022" |
| order_total	  | int	   | 总单据数         |
| order_succ	  | int	   | 成功单据数       |
| order_succ_rate | float  | 单据成功率       |
| os_total	      | int	   | 总机器数         |
| os_succ	      | int    | 成功自动交付机器数 |
| os_succ_rate	  | float  | 机器自动交付成功率 |
