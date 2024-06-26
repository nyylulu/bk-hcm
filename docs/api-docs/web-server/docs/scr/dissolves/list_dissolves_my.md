### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：我的待裁撤设备列表。

### URL

POST /api/v1/woa/dissolves/my

### 输入参数

| 参数名称     | 参数类型       | 必选 | 描述    |
|-------------|--------------|------|--------|
| operator	  | string       | 否	| 责任人  |
| module_list | string array | 否   | 模块列表 |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "module_list":["上海-南汇-M9","上海-南汇-M11","上海-南汇-M20","上海-南汇-M12","上海-南汇-M15","上海-南汇-M16","上海-南汇-M24","上海-南汇-M6","上海-南汇-M7","上海-南汇-M8"],
  "operator":"admin"
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code":0,
  "message":"OK",
  "data":{
    "total_count": 12,
    "items": [
      {
        "asset_id": "TYSV150116S1-VM7226",
        "unique_ssh_innerip": "10.205.191.116",
        "outer_ip": "NULL",
        "device_class": "D6-15-200",
        "display_name": "系统运维服务",
        "app_module": "SAWEB",
        "module_name": "上海-南汇-M9",
        "idc_unit": "上海电信南汇B15BDC2楼北2",
        "os_name": "Tencent tlinux release 1.2 (Final)",
        "input_time": "2017-01-16T00:00:00+08:00",
        "raid": "RAID5",
        "idc_logic_area": "内网云平台自营业务区1",
        "operator": "lkong",
        "bak_operator": "huibohuang"
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

| 参数名称     | 参数类型       | 描述            |
|-------------|--------------|-----------------|
| total_count | int          | 总数             |
| items       | object array | 我的待裁撤设备列表 |

#### data.items

| 参数名称            | 参数类型   | 描述         |
|--------------------|----------|--------------|
| asset_id	         | string	| 固资号        |
| unique_ssh_innerip | string	| 内网IP        |
| outer_ip	         | string	| 外网IP        |
| device_class	     | string	| SCM类型       |
| display_name	     | string	| 业务名         |
| app_module	     | string	| 业务模块       |
| module_name	     | string	| 机房Module    |
| idc_unit	         | string	| 机房 IDC单元   |
| os_name	         | string	| OS            |
| input_time	     | string	| 上架时间       |
| raid	             | string	| RAID          |
| idc_logic_area	 | string	| 逻辑区域       |
| operator	         | string	| 责任人         |
| bak_operator	     | string	| 备份责任人      |
