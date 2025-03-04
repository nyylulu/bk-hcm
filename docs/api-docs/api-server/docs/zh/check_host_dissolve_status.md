### 描述

- 该接口提供版本：v1.7.4+。
- 该接口所需权限：业务访问。
- 该接口功能描述：判断主机是否为裁撤状态。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/dissolve/hosts/status/check

### 输入参数

| 参数名称      | 参数类型      | 必选 | 描述                     |
|-----------|-----------|----|------------------------|
| bk_biz_id | int64     | 是  | 业务ID                   |
| bk_host_ids | int array | 是  | bkcc主机唯一标识列表，数组限制最大100 |


### 调用示例

```json
{
  "bk_host_ids": [11111, 22222]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "info": [
      {
        "bk_host_id": 11111,
        "status": true
      },
      {
        "bk_host_id": 22222,
        "status": false
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称 | 参数类型         | 描述       |
|------|--------------|----------|
| info | object array | 主机裁撤状态信息 |

#### info[0]

| 参数名称       | 参数类型 | 描述                                              |
|------------|------|-------------------------------------------------|
| bk_host_id | int  | bkcc主机唯一标识 |
| status     | bool | true表示主机为裁撤主机，false则不是 |
