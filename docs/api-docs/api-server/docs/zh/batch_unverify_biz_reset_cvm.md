### 描述

- 该接口提供版本：v1.8.6.0+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：批量免iOA验证的重装虚拟机接口(内部接口，只给特定业务使用，不对外公开)。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/unverify/batch/reset_async

### 输入参数

| 参数名称        | 参数类型         | 必选 | 描述                      |
|-------------|--------------|----|---------------------------------|
| bk_biz_id   | int64        | 是  | 业务ID                          |
| hosts       | object array | 是  | 虚拟机的Host列表, 最多支持500台主机 |
| pwd         | string       | 是  | 重装密码, 密码长度必须12-20位      |
| pwd_confirm | string       | 是  | 重装确认密码, 密码长度必须12-20位   |

#### hosts[n]
| 参数名称        | 参数类型    | 必选 | 描述                                          |
|----------------|-----------|-----|-----------------------------------------------|
| ip             | string	 | 是  | 主机IP                                         |
| cloud_image_id | string	 | 是  | 新镜像云ID                                      |
| image_name     | string	 | 是  | 新镜像名称                                       |

### 调用示例

```json
{
  "hosts": [
    {
      "ip": "127.0.0.1",
      "cloud_image_id": "img-002",
      "image_name": "Tencent OS 002",
    }
  ],
  "pwd": "xxxxxx",
  "pwd_confirm": "xxxxxx"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "task_management_id": "xxxxxx"
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述    |
|---------|--------|---------|
| code    | int    | 状态码   |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data参数说明

| 参数名称             | 参数类型  | 描述      |
|---------------------|---------|-----------|
| task_management_id  | string  | 任务管理id |
