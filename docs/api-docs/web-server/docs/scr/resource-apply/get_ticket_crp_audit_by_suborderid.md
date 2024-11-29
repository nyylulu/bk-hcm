### 描述

- 该接口提供版本：v1.7.0.2+。
- 该接口所需权限：平台管理-主机申领。
- 该接口功能描述：获取资源申请CRP单据审核信息。

### URL

POST /api/v1/woa/task/apply/crp_ticket/audit/get

### 输入参数

| 参数名称          | 参数类型   | 必选 | 描述       |
|---------------|--------|----|----------|
| crp_ticket_id | string | 是  | CRP单据ID  |
| suborder_id   | string | 是  | HCM子单据ID |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "crp_ticket_id": "ssss",
  "suborder_id": "111-1"
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
        "task_no": 0,
        "task_name": "部门管理员审批",
        "operate_result": "同意",
        "operator": "XXXXX",
        "operate_info": "[系统自动跳过]:审核人列表中XXXX为该单据提单人",
        "operate_time": "2024-11-04 19:57:19"
      },
      {
        "task_no": 1,
        "task_name": "业务总监审批",
        "operate_result": "同意",
        "operator": "XXXXX",
        "operate_info": "[系统自动跳过]:审核人列表中XXXX为该单据提单人",
        "operate_time": "2024-11-04 19:57:19"
      },
      {
        "task_no": 2,
        "task_name": "规划经理审批",
        "operate_result": "同意",
        "operator": "XXXXX",
        "operate_info": "[系统自动跳过]:审核人列表中XXXX为该单据提单人",
        "operate_time": "2024-11-04 19:57:19"
      },
      {
        "task_no": 3,
        "task_name": "资源经理审批",
        "operate_result": "同意",
        "operator": "XXXXX",
        "operate_info": "[系统自动跳过]:审核人列表中XXXX为该单据提单人",
        "operate_time": "2024-11-04 19:57:19"
      },
      {
        "task_no": 4,
        "task_name": "等待云上审批",
        "operate_result": "",
        "operator": "",
        "operate_info": "",
        "operate_time": ""
      },
      {
        "task_no": 5,
        "task_name": "等待交付",
        "operate_result": "",
        "operator": "",
        "operate_info": "",
        "operate_time": ""
      },
      {
        "task_no": 6,
        "task_name": "交付队列中",
        "operate_result": "",
        "operator": "",
        "operate_info": "",
        "operate_time": ""
      },
      {
        "task_no": 7,
        "task_name": "流程结束",
        "operate_result": "",
        "operator": "",
        "operate_info": "",
        "operate_time": ""
      }
    ],
    "current_step": {
      "current_task_no": 7,
      "current_task_name": "流程结束",
      "status": 129,
      "status_desc": "CVM整单创建失败",
      "fail_instance_info": [
        {
          "error_msg_type_en": "InvalidParameterValue.InsufficientOffering",
          "error_type": "下发腾讯云失败",
          "error_msg_type_cn": "云后端报资源不足",
          "request_id": "f7f83ada-d716-4dae-ad3e-b5b5ab2b9d55",
          "error_msg": "InvalidParameterValue.InsufficientOffering;RequestId=f7f83ada-d716-4dae-ad3e-b5b5ab2b9d55",
          "operator": "aaa;bbb",
          "error_count": 2
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

| 参数名称           | 参数类型   | 描述     |
|----------------|--------|--------|
| task_no        | int    | 处理节点编号 |
| task_name      | string | 处理节点名称 |
| operate_result | string | 审批结果   |
| operator       | string | 审批人    |
| operate_info   | string | 审批意见   |
| operate_time   | time   | 审批时间   |

#### current_step

| 参数名称               | 参数类型   | 描述                                                                                                                                |
|--------------------|--------|-----------------------------------------------------------------------------------------------------------------------------------|
| current_task_no    | int    | 当前任务节点编号                                                                                                                          |
| current_task_name  | string | 当前任务节点名称                                                                                                                          |
| status             | int    | 当前单据状态码。枚举值：0（部门管理员审批）,1（业务总监审批）,2（规划经理审批）,3（资源经理审批）,14（等待云上审批）,4（等待交付）,5（交付队列中）,6（资源准备中）,7（CVM生产中）,8（执行完成）,127（驳回）,129（CVM 创建失败） 
| status_desc        | string | 当前单据状态描述                                                                                                                          |
| fail_instance_info | object | 失败状态信息                                                                                                                            |

#### current_step.fail_instance_info

| 参数名称              | 参数类型   | 描述      |
|-------------------|--------|---------|
| error_msg_type_en | string | 失败状态信息  |
| error_type        | string | 错误类型    | 
| error_msg_type_cn | string | 错误信息中文  |
| error_msg         | string | 错误信息    |  
| request_id        | string | 请求 ID   | 
| operator          | string | 相关问题处理人 | 
| error_count       | int    | 错误数量    |