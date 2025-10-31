### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务访问。
- 该接口功能描述：根据回收单号查询回收设备列表。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/findmany/recycle/host

### 输入参数

| 参数名称         | 参数类型           | 必选 | 描述                             |
|--------------|------------------|------|--------------------------------|
| order_id	    | int array	      | 否   | 资源回收单号列表，数量最大20                |
| suborder_id  | string	array     | 否   | 资源回收子单号列表，数量最大20               |
| device_type  | string	array     | 否   | 机型列表，数量最大20                    |
| bk_zone_name | string array     | 否   | 地域列表，数量最大20                    |
| sub_zone     | string array     | 否   | 园区列表，数量最大20                    |
| status       | string	array     | 否   | 回收状态列表，数量最大20                  |
| bk_username  | string	array     | 否   | 提单人列表，数量最大20                   |
| ip           | string array     | 否   | 设备内网IP列表，数量最大500               |
| bk_asset_id  | string array |否 | 固资号列表，数量最大500                |
| start        | string	          | 否   | 单据创建时间过滤条件起点日期，格式如"2022-05-01" |
| end          | string	          | 否   | 单据创建时间过滤条件终点日期，格式如"2022-05-01" |
| page         | object	          | 是   | 分页信息                           |

#### page

| 参数名称      | 参数类型 | 必选 | 描述                            |
|--------------|--------|-----|---------------------------------|
| start        | int    | 否  | 记录开始位置，start 起始值为0       |
| limit        | int    | 是  | 每页限制条数，最大100              |
| enable_count | bool   | 是  | 本次请求是否为获取数量还是详情的标记  |

说明：

- enable_count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

- 默认按ip升序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "suborder_id":["1-1"],
  "page":{
    "start":0,
    "limit": 20,
    "sort":"suborder_id"
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
        "order_id":1,
        "suborder_id":"1-1",
        "bk_biz_id":2,
        "bk_biz_name":"xx",
        "bk_username":"xxx",
        "bk_asset_id":"TYSV1802949K",
        "ip":"10.0.0.1",
        "instance_id":"ins-kulm3t6z",
        "device_type":"S5.4XLARGE64",
        "bk_zone_name":"深圳",
        "sub_zone":"深圳-光明",
        "module_name":"深圳-光明-M4",
        "operator":"xxx",
        "bak_operator":"xxx",
        "input_time":"2022-08-13 16:18:39",
        "status":"DONE",
        "return_id":"TH202208150614210027",
        "return_link":"https://yunti.woa.com/orders/cvmreturn/TH202208150614210027",
        "return_tag":"计划外弹性外采购设备退回",
        "return_cost_rate":0.2,
        "return_plan_msg":"通过",
        "return_time":"2022-08-14 16:18:39",
        "create_at":"2022-08-13T15:04:05.004Z",
        "update_at":"2022-08-14T16:18:39.004Z"
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
| info    | object array | 已交付设备列表            |

#### data.info

| 参数名称       | 参数类型    | 描述          |
|---------------|-----------|---------------|
| order_id	    | int       | 资源申请单号    |
| suborder_id   | string    | 资源申请子单号  |
| bk_biz_id	    | int       | 业务ID         |
| bk_username   | string    | 提单人         |
| ip	        | string    | 设备IP         |
| asset_id      | string    | 设备固资号      |
| require_type  | int       | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤 |
| resource_type | string    | 资源类型。"QCLOUDCVM": 腾讯云虚拟机, "IDCPM": IDC物理机, "QCLOUDDVM": Qcloud富容器, "IDCDVM": IDC富容器 |
| device_type   | string    | 机型           |
| zone_name     | string    | 区域           |
| create_at     | timestamp | 记录创建时间    |
| update_at     | timestamp | 记录更新时间    |
