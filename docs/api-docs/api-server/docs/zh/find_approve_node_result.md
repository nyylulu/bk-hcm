### 描述

- 该接口提供版本：v1.8.6+。
- 该接口所需权限：无。
- 该接口功能描述：获取审批节点结果。

### URL

POST /api/v1/woa/task/find/approve_node/result

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述                    |
|-----------|--------|----|-----------------------|
| ticket_id | int    | 是  | 申请单据id                |
| state_id  | int    | 是  | 审批节点id                |


### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "ticket_id": 38829,
  "state_id": 1
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
    "name": "测试节点",
    "processed_user": "admin",
    "approve_result": true,
    "approve_remark": "快速审批，审批人：admin"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型     | 描述                         |
|---------|----------|----------------------------|
| result  | bool     | 请求成功与否。true:请求成功；false请求失败 |
| code    | int      | 错误编码。 0表示success，>0表示失败错误  |
| message | string   | 请求失败返回的错误信息                |
| data	   | object   | 响应数据                       |

#### data

| 参数名称           | 参数类型   | 描述   |
|----------------|--------|------|
| name           | string | 节点名称 |
| processed_user | string | 审批人  |
| approve_result | bool   | 审批结果 |
| approve_remark | string | 审批备注 |