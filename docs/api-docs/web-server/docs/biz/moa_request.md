### 描述

- 该接口提供版本：v1.7.0.7+。 9.9.9
- 该接口所需权限：业务访问。
- 该接口功能描述：发起MOA二次校验，仅支持对当前登录用户发起验证。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/moa/request

### 输入参数

| 参数名称     | 参数类型      | 必选 | 描述         |
|----------|-----------|----|------------|
| language | string	   | 是  | 语言:  zh/en |
| scene    | string    | 是  | 请求场景标识     |
| res_ids  | []string	 | 是  | 操作影响资源ID   |

#### scene 取值

| scene      | 操作场景  |
|------------|-------|
| sg_delete  | 安全组删除 |
| cvm_start  | CVM开机 |
| cvm_stop   | CVM关机 |
| cvm_reset  | CVM重装 |
| cvm_reboot | CVM重启 |

### 调用示例

```json
{
  "language": "zh",
  "scene": "sg_delete",
  "affected_count": 1
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

