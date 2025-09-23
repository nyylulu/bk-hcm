### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：业务视角查询资源预测单据对应的子单据。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/sub_tickets/list

### 输入参数

| 参数名称             | 参数类型         | 必选 | 描述                     |
|------------------|--------------|----|------------------------|
| ticket_id        | string       | 是  | 资源预测申请主单据ID            |
| statuses         | string array | 否  | 子单据状态列表，不传时查询全部，最多传20个 |
| sub_ticket_types | string array | 否  | 子单据类型列表，不传时查询全部，最多传20个 |
| page             | object       | 是  | 分页设置                   |

#### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                                                                                                                                    |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但不返回查询结果详情数据，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但不返回总记录条数 count                                                                                                                             |
| start | int    | 否  | 记录开始位置，start 起始值为0                                                                                                                                                                                                                                                    |
| limit | int    | 否  | 每页限制条数，最大500，不能为0                                                                                                                                                                                                                                                     |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序，默认根据submitted_at(提单时间)倒序排序，枚举值为：original_cpu_core(原始CPU核心数)、updated_cpu_core(变更后CPU核心数)、original_memory(原始内存大小)、updated_memory(变更后内存大小)、original_disk_size(原始云盘大小)、updated_disk_size(变更后云盘大小)、submitted_at(提单时间)、created_at(创建时间)、updated_at(更新时间) |
| order | string | 否  | 排序顺序，枚举值：ASC(升序)、DESC(降序)                                                                                                                                                                                                                                             |

### 调用示例

```json
{
  "ticket_id": "00000001",
  "statuses": [
    "init",
    "auditing",
    "done",
    "rejected"
  ],
  "sub_ticket_types": [
    "add",
    "adjust",
    "cancel",
    "transfer"
  ],
  "page": {
    "count": false,
    "start": 0,
    "limit": 500
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
        "id": "00000001",
        "status": "init",
        "status_name": "待审批",
        "stage": "admin_audit",
        "sub_ticket_type": "transfer",
        "ticket_type_name": "转移",
        "crp_sn": "XQ000001",
        "crp_url": "http://crp/ticket/XQ000001",
        "original_info": {
          "cvm": {
            "cpu_core": null,
            "memory": null
          }
        },
        "updated_info": {
          "cvm": {
            "cpu_core": 123,
            "memory": 123
          }
        },
        "submitted_at": "2019-07-29 11:57:20",
        "created_at": "2019-07-29 11:57:20",
        "updated_at": "2019-07-29 11:57:20"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data	   | object | 响应数据                      |

#### data

| 参数名称    | 参数类型         | 描述                                       |
|---------|--------------|------------------------------------------|
| count   | int          | 当前规则能匹配到的总记录条数，仅在 count 查询参数设置为 true 时返回 |
| details | object array | 查询返回的数据，仅在 count 查询参数设置为 false 时返回       |

#### data.details[n]

| 参数名称             | 参数类型   | 描述                                                        |
|------------------|--------|-----------------------------------------------------------|
| id               | string | 资源预测需求子单ID                                                |
| status           | string | 单据状态（枚举值：init, auditing, rejected, failed, done, invalid） |
| status_name      | string | 单据状态名称（枚举值：待审批, 审批中, 审批拒绝, 失败, 成功, 已失效）                   |
| stage            | string | 单据审批阶段（枚举值：admin_audit（部门审批）、crp_audit（公司审批））             |
| sub_ticket_type  | string | 单据类型（枚举值：add, adjust, cancel, transfer）                   |
| ticket_type_name | string | 单据类型名称                                                    |
| crp_sn           | string | CRP流程单号                                                   |
| crp_url          | string | CRP流程单链接                                                  |
| original_info    | object | 调整前的需求信息                                                  |
| updated_info     | object | 调整后的需求信息                                                  |
| submitted_at     | string | 提单时间，格式为YYYY-MM-DD HH:MM:SS，例如2024-01-01 13:59:30         |
| created_at       | string | 创建时间，格式为YYYY-MM-DD HH:MM:SS，例如2024-01-01 13:59:30         |
| updated_at       | string | 更新时间，格式为YYYY-MM-DD HH:MM:SS，例如2024-01-01 13:59:30         |

#### data.details[n].original_info & data.details[n].updated_info

| 参数名称 | 参数类型   | 描述       |
|------|--------|----------|
| cvm  | object | 申请的CVM信息 |

#### data.details[n].original_info.cvm & data.details[n].updated_info.cvm

| 参数名称     | 参数类型 | 描述       |
|----------|------|----------|
| cpu_core | int  | CPU核数（核） |
| memory   | int  | 内存总量（G）  |
