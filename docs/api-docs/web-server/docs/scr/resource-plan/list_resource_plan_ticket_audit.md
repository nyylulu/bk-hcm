### 描述

- 该接口提供版本：v1.7.0.0+。
- 该接口所需权限：平台-单据管理。
- 该接口功能描述：查询资源预测单据的审批流，包括审批状态、当前审批阶段等。

### URL

GET /api/v1/woa/plans/resources/tickets/{ticket_id}/audit

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "ticket_id": "0000000001",
    "itsm_audit": {
      "itsm_sn": "REQ000001",
      "itsm_url": "http://itsm/ticket/REQ000001",
      "status": "auditing",
      "status_name": "审批中",
      "message": "",
      "current_steps": [
        {
          "state_id": 4,
          "name": "总监审批",
          "processors": [
            "zhangsan"
          ],
          "processors_auth": {
            "zhangsan": true
          }
        }
      ],
      "logs": [
        {
          "operator": "xxx",
          "operate_at": "2024-11-06 12:00:01",
          "message": "流程开始"
        },
        {
          "operator": "xxxx",
          "operate_at": "2024-11-06 12:03:12",
          "message": "xxxx(xxxx) 处理节点【提单】(提交)"
        }
      ]
    },
    "crp_audit": {
      "crp_sn": "XQ000001",
      "crp_url": "http://crp/ticket/XQ000001",
      "status": "init",
      "status_name": "待审批",
      "message": "如果失败，这里会写原因",
      "current_steps": [
        {
          "state_id": "",
          "name": "规划经理审批",
          "processors": [
            "lisi",
            "wangwu"
          ],
          "processors_auth": {
            "lisi": true,
            "wangwu": false
          }
        }
      ],
      "logs": [
        {
          "operator": "xxxx",
          "operate_at": "2024-11-06 12:03:12",
          "message": "同意",
          "name": "部门管理员审批"
        }
      ]
    }
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

| 参数名称       | 参数类型   | 描述         |
|------------|--------|------------|
| ticket_id  | string | 资源预测需求单据ID |
| itsm_audit | object | itsm审批信息   |
| crp_audit  | object | crp审批信息    |

#### data.itsm_audit

| 参数名称          | 参数类型         | 描述                                                        |
|---------------|--------------|-----------------------------------------------------------|
| itsm_sn       | string       | ITSM流程单号                                                  |
| itsm_url      | string       | ITSM流程单链接                                                 |
| status        | string       | 审批状态（枚举值：init, auditing, rejected, done, failed, revoked） |
| status_name   | string       | 审批状态名称                                                    |
| message       | string       | 审批失败原因                                                    |
| current_steps | object array | 当前审批阶段                                                    |
| logs          | object array | 审批历史列表                                                    |

#### data.crp_audit

| 参数名称          | 参数类型         | 描述                                                        |
|---------------|--------------|-----------------------------------------------------------|
| crp_sn        | string       | CRP流程单号                                                   |
| crp_url       | string       | CRP流程单链接                                                  |
| status        | string       | 审批状态（枚举值：init, auditing, rejected, done, failed, revoked） |
| status_name   | string       | 审批状态名称                                                    |
| message       | string       | 审批失败原因                                                    |
| current_steps | object array | 当前审批阶段                                                    |
| logs          | object array | 审批历史列表                                                    |

#### data.itsm_audit.current_steps[n]

| 参数名称         | 参数类型       | 描述           |
|-----------------|--------------|----------------|
| state_id        | int          | 步骤ID         |
| name            | string       | 步骤名称        |
| processors      | string array | 审批人列表      |
| processors_auth | object       | 审批人是否有权限 |

#### data.itsm_audit.logs[n]

| 参数名称       | 参数类型   | 描述     |
|------------|--------|--------|
| operator   | string | 审批人    |
| operate_at | string | 处理时间   |
| message    | string | 审批结果信息 |

#### data.crp_audit.current_steps[n]

| 参数名称          | 参数类型      | 描述          |
|-----------------|--------------|---------------|
| state_id        | string       | 步骤ID         |
| name            | string       | 步骤名称        |
| processors      | string array | 审批人列表      |
| processors_auth | object       | 审批人是否有权限 |

#### data.crp_audit.logs[n]

| 参数名称       | 参数类型   | 描述     |
|------------|--------|--------|
| operator   | string | 审批人    |
| operate_at | string | 处理时间   |
| message    | string | 审批结果信息 |
| name       | string | 步骤名称   |
