### 描述

- 该接口提供版本：v1.8.7+。
- 该接口所需权限：无。
- 该接口功能描述：获取需要审批的itsm单据。

### URL

POST /api/v1/woa/task/apply/itsm/ticket/list

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述                |
|-------------|--------|----|-------------------|
| create_time | string | 是  |  查询创建时间大于等于该时间的单据 |


### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "create_time": "2022-11-14T01:57:41.159Z"
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
    "tickets": [
      {
        "id": "111",
        "url": "http://wwww.test.com/111",
        "user": "admin",
        "approval_state": "leader_approval",
        "create_time": "2022-11-14T01:57:41.159Z"
      }
    ]
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

| 参数名称     | 参数类型   | 描述   |
|----------|--------|------|
| tickets  | object array | 单据列表 |

#### tickets[0]

| 参数名称           | 参数类型    | 描述                                                                |
|----------------|---------|-------------------------------------------------------------------|
| id             | string  | itsm单据id                                                          |
| url            | string  | itsm链接                                                            |
| user           | string  | 提单人                                                               |
| approval_state | string  | 审批状态，"leader_approval"(直属leader审批)、"hcm_admin_approval"(hcm管理员审批) |
| create_time    | string  | 单据创建时间                                                            |
