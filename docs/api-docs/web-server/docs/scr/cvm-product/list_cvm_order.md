### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：平台-CVM生产。
- 该接口功能描述：CVM生产单据列表查询。

### URL

POST /api/v1/woa/cvm/findmany/apply/order

### 输入参数

| 参数名称         | 参数类型         | 必选 | 描述                                               |
|--------------|--------------|----|--------------------------------------------------|
| order_id	    | int array    | 否  | 资源申请单号，数量最大20                                    |
| task_id      | string array | 否  | 生产任务ID，数量最大20                                    |
| bk_username  | string array | 否  | 提单人，数量最大20                                       |
| require_type | int array	   | 否	 | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤; 6: 滚服项目; 7: 小额绿通 |
| status       | int array	   | 否	 | 单据状态。-1: 初始状态, 0: 成功, 1: 执行中, 其他: 失败             |
| region       | string array | 否	 | 地域，数量最大20                                        |
| zone 	       | string array | 否	 | 园区，数量最大20                                        |
| start        | string	      | 否  | 单据创建时间过滤条件起点日期，格式如"2022-05-01"                   |
| end          | string	      | 否  | 单据创建时间过滤条件终点日期，格式如"2022-05-01"                   |
| page         | object	      | 是  | 分页信息                                             |

#### page

| 参数名称         | 参数类型 | 必选 | 描述                 |
|--------------|------|----|--------------------|
| start        | int  | 否  | 记录开始位置，start 起始值为0 |
| limit        | int  | 是  | 每页限制条数，最大200       |
| enable_count | bool | 是  | 本次请求是否为获取数量还是详情的标记 |

**注意：**

- enable_count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

- 默认按create_at降序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id": [
    1001
  ],
  "bk_username": [
    "xxx"
  ],
  "task_id": [
    "YT000001"
  ],
  "require_type": [
    1
  ],
  "status": [
    1
  ],
  "region": [
    "ap-shanghai"
  ],
  "zone": [
    "ap-shanghai-2"
  ],
  "start": "2022-04-18",
  "end": "2022-04-25",
  "page": {
    "start": 0,
    "limit": 20,
    "enable_count": false
  }
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
    "count": 1,
    "info": [
      {
        "order_id": 1001,
        "bk_username": "admin",
        "require_type": 1,
        "remark": "",
        "spec": {
          "device_type": "S3.6XLARGE64",
          "image": "Tencent Linux Release 1.2 (tkernel2)",
          "network": "TENTHOUSAND",
          "region": "ap-shanghai",
          "zone": "ap-shanghai-2"
        },
        "task_id": "YT000001",
        "task_link": "",
        "status": 1,
        "message": "",
        "total_num": 10,
        "success_num": 5,
        "pending_num": 5,
        "success_list": [
          "10.0.0.1",
          "10.0.0.2",
          "10.0.0.3"
        ],
        "create_at": "2022-01-02T15:04:05.004Z",
        "update_at": "2022-01-02T15:04:05.004Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 响应数据                       |

#### data

| 参数名称  | 参数类型         | 描述             |
|-------|--------------|----------------|
| count | int          | 当前规则能匹配到的总记录条数 |
| info  | object array | CVM生产单据信息列表    |

#### data.info

| 参数名称         | 参数类型         | 描述                                   |
|--------------|--------------|--------------------------------------|
| order_id     | int          | 资源申请单号                               |
| bk_username  | string       | 提单人                                  |
| require_type | int	         | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤       |
| remark	      | string	      | 备注                                   |
| spec	        | object	      | 资源需求明细                               |
| task_id      | string       | 生产任务ID                               |
| task_link	   | string       | 生产任务详情链接                             |
| status	      | int          | 单据状态。-1: 初始状态, 0: 成功, 1: 执行中, 其他: 失败 |
| message	     | string       | 生产记录状态信息                             |
| total_num	   | int          | 资源需求总数                               |
| success_num  | int          | 已交付的资源数量                             |
| pending_num  | int          | 待匹配的资源数量                             |
| success_list | string array | 成功生产资源的IP列表                          |
| create_at	   | timestamp    | 步骤开始时间                               |
| update_at	   | timestamp    | 步骤结束时间                               |

**注意：**

- 如果本次请求是查询详细信息那么count为0，如果查询的是数量，那么info为空。
