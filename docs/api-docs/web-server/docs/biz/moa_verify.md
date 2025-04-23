### 描述

- 该接口提供版本：v1.7.0.7+。9.9.9
- 该接口所需权限：业务访问。
- 该接口功能描述：查询MOA二次校验结果，仅支持校验当前登录用户。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/moa/verify

### 输入参数

| 参数名称       | 参数类型      | 必选 | 描述       |
|------------|-----------|----|----------|
| scene      | string    | 是  | 请求场景标识   |
| session_id | string	   | 是	 | 会话ID     |
| res_ids    | []string	 | 是  | 操作影响资源ID |

### 调用示例

```json
{
  "scene": "sg_delete",
  "session_id": "random_string"
}
```

### 响应示例

#### session id 不存在或已过期

```json
{
  "result": false,
  "code": 2000003,
  "message": "moa session id expired or not found"
}
```

#### 验证中

```json
{
  "result": true,
  "code": 0,
  "message": "",
  "data": {
    "session_id": "random_string",
    "status": "pending"
  }
}
```

#### 验证未通过

```json
{
  "result": true,
  "code": 0,
  "message": "",
  "data": {
    "session_id": "random_string",
    "status": "finish",
    "button_type": "cancel"
  }
}
```

#### 验证通过

```json
{
  "result": true,
  "code": 0,
  "message": "",
  "data": {
    "session_id": "random_string",
    "status": "finish",
    "button_type": "confirm"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data参数说明

| 参数名称        | 参数类型   | 描述                                        |
|-------------|--------|-------------------------------------------|
| session_id  | string | 会话ID, 用于查询二次验证结果                          |
| status      | string | 二次验证状态, 枚举值: pending(验证中), finish(验证完成)   |
| button_type | string | 二次验证结果, 枚举值: confirm(验证通过), cancel(验证未通过) |

pending 验证中的状态, 前端需要持续轮询, 等到状态变为 finish 才表示验证完成。

