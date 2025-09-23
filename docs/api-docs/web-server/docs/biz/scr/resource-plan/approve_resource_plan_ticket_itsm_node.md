### 描述

- 该接口提供版本：v1.7.2+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源预测申请单据ITSM审核。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/plans/resources/tickets/{ticket_id}/approve_itsm_node

### 输入参数

| 参数名称      | 参数类型   | 必选 | 描述           |
|-----------|--------|----|--------------|
| state_id  | int	   | 是  | ITSM流程单据节点ID |
| approval	 | bool   | 是  | 是否通过         |
| remark	   | string | 否  | 审核意见         |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "state_id": 1957,
  "approval": true,
  "remark": "同意"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": null
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

无
