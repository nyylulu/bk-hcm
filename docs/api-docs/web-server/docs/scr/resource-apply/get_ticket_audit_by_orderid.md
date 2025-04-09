### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：平台管理-主机申领。
- 该接口功能描述：获取资源申请ITSM单据审核信息。

### URL

POST /api/v1/woa/task/get/apply/ticket/audit

### 输入参数

| 参数名称      | 参数类型 | 必选 | 描述   |
|-----------|------|----|------|
| order_id  | int  | 是  | 单据ID |
| bk_biz_id | int  | 是  | 业务ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id": 1001,
  "bk_biz_id": 639
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
    "order_id": 1001,
    "itsm_ticket_id": "INC202xxx21000001",
    "itsm_ticket_link": "",
    "status": "FINISHED",
    "current_steps": [
      {
        "name": "负责人审核",
        "processors": ["aaa","bbb"],
        "state_id": 1957,
        "processors_auth": {
          "aaa": true,
          "bbb": false
        }
      }
    ],
    "logs": [
      {
        "operator": "xxx",
        "operate_at": "2022-04-21 21:53:46",
        "message": "流程开始.",
        "source": "WEB"
      },
      {
        "operator": "xxx",
        "operate_at": "2022-04-21 21:53:46",
        "message": "xxx 处理【提单】",
        "source": "WEB"
      },
      {
        "operator": "xx",
        "operate_at": "2022-04-21 21:54:28",
        "message": "xxx 处理【资源管理员审核】",
        "source": "WEB"
      },
      {
        "operator": "",
        "operate_at": "2022-04-21 21:54:29",
        "message": "流程结束.",
        "source": "SYS"
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

| 参数名称             | 参数类型         | 描述                                                            |
|------------------|--------------|---------------------------------------------------------------|
| order_id         | int	         | 若order_id传值且非0，则更新order_id对应的申请单据草稿；若order_id未传值或为0，则创建申请单据草稿 |
| itsm_ticket_id   | string	      | ITSM流程单据单号                                                    |
| itsm_ticket_link | string	      | ITSM流程单据链接                                                    |
| status           | string       | 单据当前状态。"RUNNING": 处理中, "FINISHED": 已结束, "TERMINATED": 被终止     |
| current_steps	   | object array | 单据当前步骤列表                                                      |
| logs	           | object array | 单据处理日志信息列表                                                    |

#### current_steps

| 参数名称         | 参数类型       | 描述           |
|-----------------|--------------|----------------|
| name	          | string       | 步骤名称        |
| processors      | string array | 处理人列表      |
| state_id        | int          | 节点ID         |
| processors_auth | object       | 处理人是否有权限 |

#### logs

| 参数名称       | 参数类型   | 描述                                                          |
|------------|--------|-------------------------------------------------------------|
| operator   | string | 处理人                                                         |
| message    | string | 处理信息                                                        |
| operate_at | string | 处理时间                                                        |
| source     | string | 处理途径。"WEB": 页面操作, "MOBILE": 手机端操作, "API": 接口操作, "SYS": 系统操作 |
