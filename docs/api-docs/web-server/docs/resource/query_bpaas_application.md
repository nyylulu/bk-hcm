### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：资源查看。
- 该接口功能描述：查询BPaas申请。

### URL

POST /api/v1/cloud/vendors/tcloud-ziyan/applications/bpaas/query

### 输入参数

| 参数名称       | 参数类型   | 必选 | 描述   |
|------------|--------|----|------|
| bpaas_sn   | uint64 | 是  | 申请ID |
| account_id | string | 是  | 账号id |

### 调用示例

```json
{
  "bpaas_sn": 12345678,
  "account_id": "00000001"
}
```

### 响应示例

```json
 {
  "ApplicationParams": [
    {
      "Key": "key",
      "Name": "name",
      "Value": [
        "xxx"
      ]
    }
  ],
  "ApplyOwnUin": 123456789,
  "ApplyUin": 234567891,
  "ApplyUinNick": "name",
  "ApprovingNodeId": "",
  "BpaasId": 1000000001,
  "BpaasName": "bpaasname",
  "CreateTime": "2024-02-01 00:00:00",
  "ModifyTime": "2024-01-01 00:00:00",
  "Nodes": [
    {
      "ApproveId": 0,
      "ApproveMethod": -1,
      "ApproveType": -1,
      "ApprovedUin": [],
      "CKafkaRegion": "",
      "CallMethod": "",
      "CreateTime": "2024-01-01 00:00:00",
      "DataHubId": "",
      "ExternalUrl": "",
      "IsApprove": false,
      "Msg": "success",
      "NextNode": "2-3",
      "NodeId": "1",
      "NodeName": "name",
      "NodeType": 3,
      "Opinion": [],
      "ParallelNodes": "",
      "PrevNode": "0",
      "RejectedCloudFunctionMsg": "",
      "ScfName": "xxx",
      "SubStatus": 4,
      "TaskName": "",
      "Users": []
    }
  ],
  "Reason": "",
  "RequestId": "xxxxxx-7ab3-410d-xxxxx-xxxxx",
  "Status": 1
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称              | 参数类型                | 描述                              |
|-------------------|---------------------|---------------------------------|
| ApplyOwnUin       | int                 | 申请人uin                          |
| ApplyUin          | int                 | 申请人主账号                          |
| ApplyUinNick      | string              | 申请人昵称                           |
| ApprovingNodeId   | string              | 正在审批的节点id                       |
| BpaasId           | int                 | 审批流id                           |
| BpaasName         | string              | 审批流名称                           |
| Nodes             | array of StatusNode | 节点信息                            |
| Reason            | string              | 申请原因                            |
| Status            | int                 | 申请单状态 0 待审批;1 审批通过;2 拒绝;18外部审批中 |
| ApplicationParams | Array of ApplyParam | 申请参数 0 待审批;1 审批通过;2 拒绝;18外部审批中  |
| CreateTime        | string              |                                 |
| ModifyTime        | string              |                                 |
| RequestId         | string              |                                 |
