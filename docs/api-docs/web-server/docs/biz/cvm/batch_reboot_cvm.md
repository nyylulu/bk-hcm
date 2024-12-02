### 描述

- 该接口提供版本：v1.7.0.4+。
- 该接口所需权限：业务-IaaS资源操作。
- 该接口功能描述：批量重启虚拟机,异步接口&支持任务管理。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/cvms/batch/reboot_async

### 输入参数

| 参数名称       | 参数类型         | 必选 | 描述                  |
|------------|--------------|----|---------------------|
| bk_biz_id  | int64        | 是  | 业务ID                |
| ids        | string array | 是  | 虚拟机的ID列表,最多支持500个ID |
| session_id | string       | 是  | moa验证的会话ID          |


### 调用示例

```json
{
  "ids": [
    "00000001",
    "00000002"
  ],
  "session_id": "xxxxxx"
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

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |


#### data参数说明

| 参数名称               | 参数类型   | 描述     |
|--------------------|--------|--------|
| task_management_id | string | 任务管理id |
