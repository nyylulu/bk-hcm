### 描述

- 该接口提供版本：v1.4.0+。
- 该接口所需权限：IaaS资源操作。
- 该接口功能描述：根据CVM cloud_id 关联安全组，安全组为本地id (目前仅支持自研云)。

### URL

POST /api/v1/cloud/bizs/{bk_biz_id}/security_groups/associate/cloud_cvms/batch

### 输入参数

| 参数名称              | 参数类型         | 必选 | 描述      |
|-------------------|--------------|----|---------|
| bk_biz_id         | int64        | 是  | 业务id    |
| security_group_id | string       | 是  | 安全组ID   |
| cloud_cvm_ids     | string array | 是  | 主机云ID列表 |

### 调用示例

```json
{
  "security_group_id": "00001111",
  "cloud_cvm_ids": ["ins-xxx"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok"
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
