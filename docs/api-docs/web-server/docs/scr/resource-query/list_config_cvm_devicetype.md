### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：无。
- 该接口功能描述：机型配置信息查询。

### URL

POST /api/v1/woa/config/findmany/config/cvm/devicetype

### 输入参数

| 参数名称   | 参数类型   | 必选 | 描述     |
|--------|--------|----|--------|
| filter | object | 是  | 查询过滤条件 |

#### filter

| 参数名称      | 参数类型        | 必选 | 描述                                                              |
|-----------|-------------|----|-----------------------------------------------------------------|
| condition | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rule      | array       | 是  | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rule[n] （详情请看 rules 表达式说明）

| 参数名称     | 参数类型        | 必选 | 描述                                          |
|----------|-------------|----|---------------------------------------------|
| field    | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| operator | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value    | 可变类型        | 是  | 查询条件Value值                                  |

##### rule 表达式说明：

##### 1. 操作符

| 操作符              | 描述                                        | 操作符的value支持的数据类型                              |
|------------------|-------------------------------------------|-----------------------------------------------|
| equal            | 等于。不能为空字符串                                | boolean, numeric, string                      |
| not_equal        | 不等。不能为空字符串                                | boolean, numeric, string                      |
| greater          | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| greater_or_equal | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| less             | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| less_or_equal    | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in               | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                      |
| not_in           | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                      |
| contains         | 模糊查询，区分大小写                                | string                                        |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "filter": {
    "condition": "AND",
    "rules": [
      {
        "field": "zone",
        "operator": "equel",
        "value": "ap-shanghai-2"
      },
      {
        "field": "cpu",
        "operator": "equel",
        "value": 4
      }
    ]
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "count": 2,
    "info": [
      {
        "device_type": "S3.MEDIUM4",
        "device_type_class": "CommonType",
        "cpu_amount": 2,
        "ram_amount": 16
      },
      {
        "device_type": "S3.LARGE8",
        "device_type_class": "CommonType",
        "cpu_amount": 4,
        "ram_amount": 32
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 响应数据                       |

#### data

| 参数名称  | 参数类型         | 描述             |
|-------|--------------|----------------|
| count | int          | 当前规则能匹配到的总记录条数 |
| info  | object array | 机型列表           |

#### info[0]

| 参数名称              | 参数类型   | 描述                                |
|-------------------|--------|-----------------------------------|
| device_type       | string | 机型                                |
| device_type_class | string | 通/专用机型，SpecialType专用，CommonType通用 |
| device_group      | string | 机型族                               |
| cpu_amount        | int    | cpu核数                             |
| ram_amount        | int    | 内存容量                             |
| core_type         | int    | 核心类型(1:小核心 2:中核心 3:大核心)     |
| device_class      | string | 实例类型                             |
