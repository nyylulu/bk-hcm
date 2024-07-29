### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：业务-IaaS资源删除。
- 该接口功能描述：资源可回收状态检查。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/findmany/recycle/recyclability

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述        |
|--------------|--------------|------|------------|
| ips     	   | string array | 否   | 要查询的ip列表，数量最大500 |
| asset_ids    | string array | 否   | 要查询的固资号列表，数量最大500      |
| bk_host_ids  | int array    | 否   | 要查询的CC主机ID，数量最大500 |

说明：

- ips、asset_ids和bk_host_ids不能同时为空

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "ips": [
    "10.0.0.1"
  ]
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
        "bk_host_id":17,
        "asset_id":"TYSV1802949K",
        "ip":"10.0.0.1",
        "outer_ip": "10.0.0.2",
        "bk_biz_id":2,
        "bk_biz_name":"",
        "topo_module":"待回收",
        "operator":"xx",
        "bak_operator":"xx",
        "device_type":"D4-8-100-10",
        "state":"开发使用中[无告警]",
        "input_time":"2020-07-03 00:00:00",
        "recyclable":false,
        "message":"必须为主机负责人或备份负责人;主机模块不是[待回收模块.待回收]"
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
| info    | object array | 可回收状态信息列表         |

#### data.info

| 参数名称      | 参数类型  | 描述          |
|--------------|---------|---------------|
| bk_host_id   | int	 | CC主机ID       |
| asset_id	   | string	 | 设备固资号      |
| ip	       | string	 | 设备内网ip     |
| outer_ip	   | string	 | 设备外网ip     |
| bk_biz_id	   | int	 | CC业务ID       |
| bk_biz_name  | string	 | CC业务名       |
| topo_module  | string	 | 模块名         |
| operator     | string	 | 主机负责人      |
| bak_operator | string	 | 主机备份负责人   |
| device_type  | string	 | 机型           |
| state	       | string  | 主机状态        |
| input_time   | string	 | 入库时间        |
| recyclable   | bool	 | 可回收状态      |
| message      | string	 | 回收状态详细信息 |
