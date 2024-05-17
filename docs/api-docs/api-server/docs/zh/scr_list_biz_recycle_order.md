### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询业务资源回收历史单据信息。

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述        |
|--------------|--------------|------|------------|
| bk_biz_id    | int          | 是   | 业务ID      |
| start        | string       | 否   | 单据创建时间过滤条件起点日期，格式如"2022-05-01" |
| end          | string       | 否   | 单据创建时间过滤条件终点日期，格式如"2022-05-01"  |
| page         | object	      | 是   | 分页信息     |

#### page

| 参数名称      | 参数类型 | 必选 | 描述                            |
|--------------|--------|-----|---------------------------------|
| start        | int    | 否  | 记录开始位置，start 起始值为0       |
| limit        | int    | 是  | 每页限制条数，最大200              |
| enable_count | bool   | 是  | 本次请求是否为获取数量还是详情的标记  |

注意：

- 默认按create_at降序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "bk_biz_id":2,
  "start":"2022-01-01",
  "end":"2022-01-07",
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
  "permission":null,
  "request_id":"f5a6331d4bc2433587a63390c76ba7bf",
  "data":{
    "count":1,
    "info":[
      {
        "order_id":1,
        "suborder_id":"1-1",
        "bk_biz_id":2,
        "bk_biz_name":"xx",
        "bk_username":"xx",
        "resource_type":"IDCPM",
        "recycle_type":"常规项目",
        "return_plan":"IMMEDIATE",
        "pool_type":0,
        "cost_concerned":true,
        "stage":"DONE",
        "status":"DONE",
        "message":"",
        "handler":"AUTO",
        "total_num":1,
        "success_num":1,
        "pending_num":0,
        "failed_num":0,
        "remark":"",
        "create_at":"2023-03-15T09:39:56.763Z",
        "update_at":"2023-03-15T09:40:41.931Z"
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
| permission | object       | 权限信息             |
| request_id | string       | 请求链ID             |
| data	     | object array | 响应数据             |

#### data

| 参数名称 | 参数类型       | 描述                    |
|---------|--------------|-------------------------|
| count   | int          | 当前规则能匹配到的总记录条数 |
| info    | object array | 资源申请单据信息列表       |

#### info 字段说明：
| 参数名称              | 参数类型   | 描述                         |
|---------------------|-----------|------------------------------|
| order_id            | int       | 资源申请单号 |
| suborder_id         | string    | 资源申请子单号 |
| bk_biz_id           | int       | 业务ID |
| bk_username         | string    | 提单人 |
| resource_type       | string    | 资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "OTHERS": 其他 |
| recycle_type        | string    | 回收类型。"常规项目", "机房裁撤", "过保裁撤", "不区分" |
| return_plan         | string    | 退回策略。"IMMEDIATE": 立即销毁, "DELAY": 延迟销毁 |
| cost_concerned      | bool      | 是否涉及回收成本 |
| stage               | string    | 单据执行阶段。"UNCOMMIT": 未提交, "AUDIT": 审核中, "RUNNING": 生产中, "SUSPEND": 备货异常, "DONE": 已完成 |
| status              | string    | 单据状态。"WAIT": 待匹配, "MATCHING": 匹配执行中, "MATCHED_SOME": 已完成部分资源匹配, "PAUSED": 已暂停, "DONE": 完成, "TERMINATE": 终止 |
| message             | string    | 单据状态说明 |
| handler             | string    | 当前处理人 |
| total_num           | int       | 资源需求总数 |
| success_num         | int       | 已交付的资源数量 |
| pending_num         | int       | 待匹配的资源数量 |
| failed_num          | int       | 回收失败的资源数量 |
| remark              | string    | 备注 |
| start_at            | timestamp | 步骤开始时间 |
| end_at              | timestamp | 步骤结束时间 |
