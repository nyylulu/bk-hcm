### 描述

- 该接口提供版本：v1.7.4+。
- 该接口所需权限：业务访问。
- 该接口功能描述：检查主机是否有未完结的uwork单据。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/hosts/uwork_tickets/status/check

### 输入参数

| 参数名称        | 参数类型       | 必选 | 描述                |
|-------------|------------|----|-------------------|
| bk_biz_id   | int64      | 是  | 业务ID              |
| bk_host_ids | int array	 | 是	 | 要查询的主机ID列表，最大100条 |

### 调用示例

```json
{
  "bk_host_ids": [
    11111,
    22222
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "details": [
      {
        "bk_host_id": 11111,
        "has_open_tickets": true,
        "open_ticket_ids": [
          "111",
          "222",
          "333(process_name)"
        ]
      },
      {
        "bk_host_id": 22222,
        "has_open_tickets": false
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data	   | object | 响应数据                      |

#### data

| 参数名称    | 参数类型         | 描述     |
|---------|--------------|--------|
| details | object array | 响应数据数组 |

#### details

| 参数名称              | 参数类型         | 描述                                |
|-------------------|--------------|-----------------------------------|
| bk_host_id	       | int          | 主机ID                              |
| has_open_tickets	 | bool         | 是否有未完结的单据                         |
| open_ticket_ids	  | string array | 未完结的流程单号，故障单只有单号，其他流程单会附带流程名（如果有） |
