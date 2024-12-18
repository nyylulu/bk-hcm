### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：资源回收单据列表查询。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/findmany/recycle/order

### 输入参数

| 参数名称       | 参数类型          | 必选 | 描述        |
|-------------- |-----------------|------|------------|
| order_id	    | array int	      | 否   | 资源申请单号 |
| suborder_id   | array string    | 否   | 资源申请子单号  |
| stage         | array string    | 否   | 状态(枚举值:COMMIT、DETECT、AUDIT、TRANSIT、RETURN、DONE、TERMINATE) |
| resource_type | array string    | 否   | 需求资源类型(枚举值:"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "OTHERS": 其他) |
| recycle_type  | array string    | 否   | 回收类型(枚举值:常规项目、机房裁撤、过保裁撤、春节保障、滚服项目)  |
| return_plan   | array string    | 否   | 退回策略(枚举值:"IMMEDIATE": 立即销毁, "DELAY": 延迟销毁)  |
| bk_username   | string	      | 否   | 提单人      |
| start         | string	      | 否   | 单据创建时间过滤条件起点日期，格式如"2022-05-01" |
| end           | string	      | 否   | 单据创建时间过滤条件终点日期，格式如"2022-05-01" |
| page          | object	      | 是   | 分页信息     |

#### page

| 参数名称      | 参数类型 | 必选 | 描述                            |
|--------------|--------|-----|---------------------------------|
| start        | int    | 否  | 记录开始位置，start 起始值为0       |
| limit        | int    | 是  | 每页限制条数，最大200              |
| enable_count | bool   | 是  | 本次请求是否为获取数量还是详情的标记  |

说明：

- enable_count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

- 默认按create_at降序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id":[1001],
  "bk_username":["xxx"],
  "start":"2022-04-18",
  "end":"2022-04-25",
  "page":{
    "start":0,
    "limit":20,
    "enable_count":false
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
        "order_id":1001,
        "bk_biz_id":2,
        "bk_username":"xxx",
        "status":"DONE",
        "total_num":10,
        "success_num":5,
        "pending_num":0,
        "failed_num":5,
        "remark":"",
        "create_at":"2022-01-02T15:04:05.004Z",
        "update_at":"2022-01-02T15:04:05.004Z"
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
| info    | object array | 资源回收单据信息列表       |

#### data.info

| 参数名称             | 参数类型    | 描述            |
|---------------------|-----------|-----------------|
| order_id            | int       | 资源申请单号      |
| bk_biz_id	          | int	      | 业务ID           |
| bk_username         |	string    | 提单人           |
| status              |	string    | 单据状态。INIT：待回收，RUNNING：回收执行中，PAUSED：已暂停，DONE：完成 |
| total_num           |	int	      | 回收资源总数      |
| success_num         |	int	      | 回收成功的资源数量 |
| pending_num         |	int	      | 待回收的资源数量   |
| failed_num          |	int	      | 回收失败的资源数量 |
| remark	          | string	  | 备注             |
| start_at	          | timestamp |	步骤开始时间      |
| end_at	          | timestamp |	步骤结束时间      |

**注意：**

- 如果本次请求是查询详细信息那么count为0，如果查询的是数量，那么info为空。