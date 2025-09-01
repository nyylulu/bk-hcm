### 描述

- 该接口提供版本：v1.8.5.1+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：蓝鲸审批助手-用户确认操作的回调接口。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/confirm/apply/record/modify

### 输入参数

| 参数名称        | 参数类型       | 必选   | 描述                                             |
|----------------|--------------|--------|-------------------------------------------------|
| bk_biz_id      | int	        | 是	 | 业务ID                                           |
| bk_username    | string       | 是	 | 资源申请提单人                                     |
| action	     | string       | 是	 | 按钮标识(APPROVE:通过 REJECT:拒绝)                 |
| suborder_id	 | string	    | 是	 | 主机申请子单号                                     |
| modify_id      | int	        | 是	 | 主机申请变更单号                                   |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "bk_username": "xx",
  "action": "APPROVE",
  "suborder_id": "1001_1",
  "modify_id": 1001
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
    "response_msg": "单据确认成功",
    "response_color": "green"
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

| 参数名称         | 参数类型 | 描述                              |
|----------------|---------|-----------------------------------|
| response_msg   | string  | 点击按钮后回显的信息                 |
| response_color | string  | 点击按钮后显示的颜色，只支持green、red |
