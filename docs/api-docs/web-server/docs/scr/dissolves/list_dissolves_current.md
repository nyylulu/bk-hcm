### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：当前列表。

### URL

POST /api/v1/woa/dissolves/current

### 输入参数

### 输入参数

| 参数名称      | 参数类型       | 必选 | 描述     |
|--------------|--------------|------|---------|
| display_name | string       | 是   | 业务名称  |
| module_list  | string array | 是   | 模块列表  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "display_name":"蓝鲸运营",
  "module_list":["广州-人民中-M1","美国-加州-M2","广州-亚太-M3","广州-亚太-M1"]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code":0,
  "message":"OK",
  "data":{
    "total_count": 6,
    "items": [
      {
        "asset_id": "TYSV11081FEC-4",
        "inner_ip": "10.204.185.38",
        "outer_ip": "NULL",
        "display_name": "蓝鲸运营",
        "app_module": "nfs",
        "device_class": "C1",
        "module_name": "广州-人民中-M1",
        "idc_unit": "广州电信人民中路AC7楼",
        "os_name": "Tencent tlinux release 2.0 (Final)",
        "input_time": "2012-06-29 00:00:00",
        "raid": "NORAID",
        "idc_logic_area": "合作业务区1",
        "operator": "admin",
        "bak_operator": "admin"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object array | 响应数据             |

#### data

| 参数名称     | 参数类型       | 描述        |
|-------------|--------------|-------------|
| total_count | int          | 总数         |
| items       | object array | 机器列表     |

#### data.items

| 参数名称         | 参数类型   | 描述             |
|-----------------|----------|------------------|
| asset_id	      | string	 | 固资编号          |
| inner_ip	      | string	 | 内网IP           |
| outer_ip	      | string	 | 公网IP           |
| display_name	  | string	 | 业务名称          |
| app_module	  | string	 | 业务模块          |
| device_class	  | string	 | SCM设备类型       |
| module_name	  | string	 | ModuleName       |
| idc_unit	      | string	 | 存放机房管理单元    |
| os_name	      | string	 | 操作系统           |
| input_time	  | string	 | 上架时间           |
| raid	          | string	 | RAID结构          |
| idc_logic_area  | string   | 逻辑区域           |
| operator	      | string	 | 维护人             |
| bak_operator	  | string	 | 备份维护人          |
