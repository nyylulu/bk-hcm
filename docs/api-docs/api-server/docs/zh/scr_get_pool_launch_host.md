### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：平台-资源上下架。
- 该接口功能描述：查询资源池上架设备。

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述        |
|--------------|--------------|------|------------|
| id           | int          | 是   | 上架任务ID   |
| page         | object	      | 是   | 分页信息     |

#### page

| 参数名称      | 参数类型 | 必选 | 描述                            |
|--------------|--------|-----|---------------------------------|
| start        | int    | 否  | 记录开始位置，start 起始值为0       |
| limit        | int    | 是  | 每页限制条数，最大200              |
| enable_count | bool   | 是  | 本次请求是否为获取数量还是详情的标记  |

注意：

- enable_count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

- 默认按id升序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "id":1,
  "page":{
    "start":0,
    "limit":10
  }
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data":{
    "count":1,
    "info":[
      {
        "bk_host_id":124935,
        "labels":{
          "ip":"10.0.0.1",
          "bk_asset_id":"TYSVxxx",
          "resource_type":"IDCPM",
          "device_type":"CG2-10G",
          "region":"南京",
          "zone":"南京-吉山",
          "grade_tag":"D3"
        },
        "op_type":"LAUNCH",
        "task_id":1,
        "phase":"SUCCESS",
        "message":"",
        "operator":"xxx",
        "create_at":"2022-11-14T01:57:41.159Z",
        "update_at":"2022-11-14T01:57:41.159Z"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 上架设备信息列表           |

#### info 字段说明：

| 参数名称     | 参数类型           | 描述                             |
|------------|-------------------|----------------------------------|
| bk_host_id | int               | CC主机ID                          |
| labels     | map[string]string | 设备标签                           |
| op_type    | string            | 操作类型。LAUNCH：上架，RECALL：下架 |
| task_id    | int               | 关联任务ID                         |
| phase      | string            | 执行阶段。INIT：待执行，RUNNING：执行中，PAUSED：已暂停，SUCCESS：执行成功，FAILED：执行失败 |
| message    | string            | 执行信息                           |
| operator   | string            | 操作者                             |
| create_at  | timestamp         | 记录创建时间                        |
| update_at  | timestamp         | 记录最后更新时间                     |

**注意：**
- 如果本次请求是查询详细信息那么count为0，如果查询的是数量，那么info为空。
