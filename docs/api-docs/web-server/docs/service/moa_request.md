### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：无。
- 该接口功能描述：发起MOA二次校验。

### URL

POST /api/v1/web/moa/request

### 输入参数

| 参数名称           | 参数类型    | 必选 | 描述                                        |
|----------------|---------|----|-------------------------------------------|
| username	      | string	 | 是	 | 用户名                                       |
| channel	       | string	 | 是	 | 	使用哪种二次验证通道(1) moa: MOA弹窗确认(2) sms: 短信验证码 |
| language	      | string	 | 是	 | 语言                                        |
| prompt_payload | string  | 是	 | 二次验证弹窗内容, 参考调用示例进行调整                      |


### 调用示例


```json
{
    "username": "zhangsan",
    "channel": "moa",
    "language": "zh",
    "prompt_payload": "{\"zh\":{\"title\":\"新设备登录授权\",\"navigator\":\"导航栏\",\"desc\":\"您的账号正在新设备登录MOA，是否同意本次操作？\",\"footer\":\"\",\"buttons\":[{\"desc\":\"确定\",\"button_type\":\"confirm\"},{\"desc\":\"取消\",\"button_type\":\"cancel\"}]},\"en\":{\"title\":\"Two-step verification\",\"navigator\":\"navigator\",\"desc\":\"A new device is signing in MOA\",\"icon_url\":\"https://xxx.xxx.xxx\",\"footer\":\"\",\"buttons\":[{\"desc\":\"Allow\",\"button_type\":\"confirm\"},{\"desc\":\"Do Not Allow\",\"button_type\":\"cancel\"}]}}"
}
```

### 响应示例

```json
{
  "result": true,
  "code": 0,
  "message": "",
  "data": {
    "session_id": "random_string"
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
| 参数名称       | 参数类型   | 描述               |
|------------|--------|------------------|
| session_id | string | 会话ID, 用于查询二次验证结果 |

