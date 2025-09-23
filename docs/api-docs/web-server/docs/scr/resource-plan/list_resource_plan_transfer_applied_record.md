### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：平台-资源预测。
- 该接口功能描述：查询资源下资源预测转移额度执行记录列表。

### URL

POST /api/v1/woa/plans/resources/transfer_applied_records/list

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述        |
|-----------|--------|------|------------|
| filter    | object | 是   | 查询过滤条件  |
| page      | object | 是   | 分页设置     |

#### filter

| 参数名称  | 参数类型        | 必选 | 描述                                                              |
|-------|-------------|----|-----------------------------------------------------------------|
| op    | enum string | 是  | 操作符（枚举值：and、or）。如果是and，则表示多个rule之间是且的关系；如果是or，则表示多个rule之间是或的关系。 |
| rules | array       | 是  | 过滤规则，最多设置5个rules。如果rules为空数组，op（操作符）将没有作用，代表查询全部数据。             |

#### rules[n] （详情请看 rules 表达式说明）

| 参数名称  | 参数类型        | 必选 | 描述                                          |
|-------|-------------|----|---------------------------------------------|
| field | string      | 是  | 查询条件Field名称，具体可使用的用于查询的字段及其说明请看下面 - 查询参数介绍  |
| op    | enum string | 是  | 操作符（枚举值：eq、neq、gt、gte、le、lte、in、nin、cs、cis） |
| value | 可变类型        | 是  | 查询条件Value值                                  |

##### rules 表达式说明：

##### 1. 操作符

| 操作符 | 描述                                        | 操作符的value支持的数据类型                              |
|-----|-------------------------------------------|-----------------------------------------------|
| eq  | 等于。不能为空字符串                                | boolean, numeric, string                      |
| neq | 不等。不能为空字符串                                | boolean, numeric, string                      |
| gt  | 大于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| gte | 大于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lt  | 小于                                        | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| lte | 小于等于                                      | numeric，时间类型为字符串（标准格式："2006-01-02T15:04:05Z"） |
| in  | 在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素  | boolean, numeric, string                      |
| nin | 不在给定的数组范围中。value数组中的元素最多设置100个，数组中至少有一个元素 | boolean, numeric, string                      |
| cs  | 模糊查询，区分大小写                                | string                                        |
| cis | 模糊查询，不区分大小写                               | string                                        |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                                        |
|-------|--------|----|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count                                 |
| start | int    | 否  | 记录开始位置，start 起始值为0                                                                                                                                                        |
| limit | int    | 否  | 每页限制条数，最大500，不能为0                                                                                                                                                         |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序，默认根据submitted_at(提单时间)倒序排序，枚举值为：cpu_core(CPU核心数)、memory(内存大小)、disk_size(云盘大小)、expect_time(期望交付时间)、submitted_at(提单时间)、created_at(创建时间)、updated_at(更新时间) |
| order | string | 否  | 排序顺序，枚举值：ASC(升序)、DESC(降序)                                                                                                                                                 |

### 调用示例

```json
{
  "op": "and",
  "rules": [
    {
      "field": "name",
      "op": "eq",
      "value": "Jim"
    },
    {
      "field": "age",
      "op": "gt",
      "value": 18
    },
    {
      "field": "age",
      "op": "lt",
      "value": 30
    }
  ],
  "page": {
    "count": false,
    "start": 0,
    "limit": 100
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "details": [
      {
        "id": "0000001z",
        "applied_type": "add",
        "bk_biz_id": 1001,
        "sub_ticket_id": "100001_1",
        "year": 2025,
        "technical_class": "标准型",
        "obs_project": "常规项目",
        "expected_core": 100,
        "applied_core": 200,
        "creator": "Jim",
        "reviser": "Jim",
        "created_at": "2023-02-12T14:47:39Z",
        "updated_at": "2023-02-12T14:55:40Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型   | 描述                               |
|---------|-------- --|------------------------------------|
| code    | int       | 错误编码。 0表示success，>0表示失败错误 |
| message | string    | 请求失败返回的错误信息                 |
| data	  | object    | 响应数据                             |

#### data

| 参数名称 | 参数类型 | 描述                    |
|---------|--------|-------------------------|
| count   | int    | 当前规则能匹配到的总记录条数 |
| details | array  | 查询返回的数据             |

#### data.details[n]

| 参数名称            | 参数类型       | 描述                                   |
|--------------------|--------------|----------------------------------------|
| id                 | string       | 资源ID                                 |
| applied_type       | string       | 转移类型（枚举值：add(转移进池)、remove(转移出池)）|
| bk_biz_id          | int          | 业务ID                                 |
| sub_ticket_id      | string       | 预测调整子单号                           |
| year               | int          | 额度所属年份                             |
| technical_class    | string       | 技术分类                                |
| obs_project        | string       | 项目类型                                |
| expected_core      | int          | 预期转移的核心数                          |
| applied_core       | int          | 成功转移的核心数                          |
| creator            | string       | 创建者                                   |
| reviser            | string       | 修改者                                   |
| created_at         | string       | 创建时间，标准格式：2006-01-02T15:04:05Z   |
| updated_at         | string       | 修改时间，标准格式：2006-01-02T15:04:05Z   |
