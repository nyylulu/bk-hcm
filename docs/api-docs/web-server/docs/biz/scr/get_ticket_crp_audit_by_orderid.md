### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：获取资源申请CRP单据审核信息。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/apply/crp_ticket/audit/get

### 输入参数

| 参数名称          | 参数类型   | 必选 | 描述      |
|---------------|--------|----|---------|
| crp_ticket_id | string | 是  | CRP单据ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "crp_ticket_id": "ssss"
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
    "crp_ticket_id": "INC202xxx21000001",
    "crp_ticket_link": "",
    "logs": [
      {
        "taskNo": "0",
        "taskName": "部门管理员审批",
        "operateResult": "同意",
        "operator": "XXXXX",
        "operateInfo": "[系统自动跳过]:审核人列表中XXXX为该单据提单人",
        "operateTime": "2024-11-04 19:57:19"
      },
      {
        "taskNo": "1",
        "taskName": "业务总监审批",
        "operateResult": "同意",
        "operator": "XXXXX",
        "operateInfo": "[系统自动跳过]:审核人列表中XXXX为该单据提单人",
        "operateTime": "2024-11-04 19:57:19"
      },
      {
        "taskNo": "2",
        "taskName": "规划经理审批",
        "operateResult": "同意",
        "operator": "XXXXX",
        "operateInfo": "[系统自动跳过]:审核人列表中XXXX为该单据提单人",
        "operateTime": "2024-11-04 19:57:19"
      },
      {
        "taskNo": "3",
        "taskName": "资源经理审批",
        "operateResult": "同意",
        "operator": "XXXXX",
        "operateInfo": "[系统自动跳过]:审核人列表中XXXX为该单据提单人",
        "operateTime": "2024-11-04 19:57:19"
      },
      {
        "taskNo": "4",
        "taskName": "等待云上审批",
        "operateResult": "",
        "operator": "",
        "operateInfo": "",
        "operateTime": ""
      },
      {
        "taskNo": "5",
        "taskName": "等待交付",
        "operateResult": "",
        "operator": "",
        "operateInfo": "",
        "operateTime": ""
      },
      {
        "taskNo": "6",
        "taskName": "交付队列中",
        "operateResult": "",
        "operator": "",
        "operateInfo": "",
        "operateTime": ""
      },
      {
        "taskNo": "7",
        "taskName": "流程结束",
        "operateResult": "",
        "operator": "",
        "operateInfo": "",
        "operateTime": ""
      }
    ],
    "current_step": {
      "currentTaskNo": 7,
      "currentTaskName": "流程结束",
      "status": 129,
      "statusDesc": "CVM整单创建失败",
      "failInstanceInfo": [
        {
          "errorMsgTypeEn": "InvalidParameterValue.InsufficientOffering",
          "errorType": "下发腾讯云失败",
          "errorMsgTypeCn": "云后端报资源不足",
          "requestId": "f7f83ada-d716-4dae-ad3e-b5b5ab2b9d55",
          "errorMsg": "InvalidParameterValue.InsufficientOffering;RequestId=f7f83ada-d716-4dae-ad3e-b5b5ab2b9d55",
          "operator": "cutechen;lotuschen",
          "errorCount": 2
        }
      ]
    }
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                         |
|---------|--------|----------------------------|
| result  | bool   | 请求成功与否。true:请求成功；false请求失败 |
| code    | int    | 错误编码。 0表示success，>0表示失败错误  |
| message | string | 请求失败返回的错误信息                |
| data	   | object | 响应数据                       |

#### data

| 参数名称            | 参数类型         | 描述         |
|-----------------|--------------|------------|
| crp_ticket_id   | string	      | CRP流程单据单号  |
| crp_ticket_link | string	      | CRP流程单据链接  |
| logs	           | object array | 单据处理日志信息列表 |
| current_step    | object       | 单据当前步骤列表   |

#### logs

| 参数名称          | 参数类型   | 描述     |
|---------------|--------|--------|
| taskNo        | int    | 处理节点编号 |
| taskName      | string | 处理节点名称 |
| operateResult | string | 审批结果   |
| operator      | string | 审批人    |
| operateInfo   | string | 审批意见   |
| operateTime   | time   | 审批时间   |

#### current_step

| 参数名称             | 参数类型   | 描述                                                                                                                                |
|------------------|--------|-----------------------------------------------------------------------------------------------------------------------------------|
| currentTaskNo    | int    | 当前任务节点编号                                                                                                                          |
| currentTaskName  | string | 当前任务节点名称                                                                                                                          |
| status           | int    | 当前单据状态码。枚举值：0（部门管理员审批）,1（业务总监审批）,2（规划经理审批）,3（资源经理审批）,14（等待云上审批）,4（等待交付）,5（交付队列中）,6（资源准备中）,7（CVM生产中）,8（执行完成）,127（驳回）,129（CVM 创建失败） 
| statusDesc       | string | 当前单据状态描述                                                                                                                          |
| failInstanceInfo | object | 失败状态信息                                                                                                                            |

#### current_step.failInstanceInfo

| 参数名称           | 参数类型   | 描述      |
|----------------|--------|---------|
| errorMsgTypeEn | string | 失败状态信息  |
| errorType      | string | 错误类型    | 
| errorMsgTypeCn | string | 错误信息中文  |
| errorMsg       | string | 错误信息    |  
| requestId      | string | 请求 ID   | 
| operator       | string | 相关问题处理人 | 
| errorCount     | int    | 错误数量    |