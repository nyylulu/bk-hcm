### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：平台-CVM生产。
- 该接口功能描述：CVM实例查询。

### URL

POST /api/v1/woa/cvm/findmany/apply/device

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述        |
|--------------|--------------|------|------------|
| order_id	   | int array    | 否   | 资源申请单号，数量最大20 |
| page         | object	      | 是   | 分页信息     |

#### page

| 参数名称      | 参数类型 | 必选 | 描述                            |
|--------------|--------|-----|---------------------------------|
| start        | int    | 否  | 记录开始位置，start 起始值为0       |
| limit        | int    | 是  | 每页限制条数，最大200              |
| enable_count | bool   | 是  | 本次请求是否为获取数量还是详情的标记  |

**注意：**

enable_count 如果此标记为true，表示此次请求是获取数量。此时其余字段必须为初始化值，start为0,limit为:0。

默认按create_at降序排序

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "order_id":1001,
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
        "asset_id":"TC220622003911",
        "ip": "10.0.0.1",
        "cvm_inst_id":"ins-8d6qs59o"
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
| info    | object array | CVM实例信息列表           |

#### data.info

| 参数名称      | 参数类型    | 描述       |
|--------------|-----------|-----------|
| asset_id     | string    | 设备固资号  |
| ip           | string	   | 设备IP     |
| cvm_inst_id  | string	   | CVM实例ID  |

**注意：**

- 如果本次请求是查询详细信息那么count为0，如果查询的是数量，那么info为空。
