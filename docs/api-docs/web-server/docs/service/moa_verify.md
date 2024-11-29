### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：查询MOA二次校验结果。

### URL

POST /api/v1/web/moa/verify

### 输入参数

| 参数名称        | 参数类型    | 必选 | 描述                   |
|-------------|---------|----|----------------------|
| username	   | string	 | 是	 | 用户名                  |
| session_id	 | string	 | 是	 | 	会话ID                |



### 调用示例


```json
{
  "username": "zhangsan",
  "session_id": "random_string"
}
```

### 响应示例

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

