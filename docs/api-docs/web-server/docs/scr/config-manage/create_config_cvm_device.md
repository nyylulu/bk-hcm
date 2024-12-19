### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：平台-CVM机型。
- 该接口功能描述：CVM机型配置信息创建。

### URL

POST /api/v1/woa/config/createmany/config/cvm/device

### 输入参数

| 参数名称       | 参数类型       | 必选 | 描述          |
|---------------|--------------|------|--------------|
| require_type	| int array	   | 是	  | 需求类型       |
| zone	        | string array | 是	  | 可用区         |
| device_group	| string	   | 是   | 机型族         |
| device_size	| string	   | 是   | 核心类型，枚举值：小核心、中核心、大核心 |
| device_type	| string	   | 是   | 设备型号        |
| cpu	        | int	       | 是   | CPU核数，单位个  |
| mem	        | int	       | 是   | 内存大小，单位G  |
| remark        | string	   | 否   | 其他信息	      |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "require_type":[
    1
  ],
  "zone":[
    "ap-shanghai-2"
  ],
  "device_group":"标准型",
  "device_size":"大核心",
  "device_type":"S2.LARGE16",
  "cpu":4,
  "mem":16,
  "remark":""
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result":true,
  "code":0,
  "message":"success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object       | 请求返回的数据        |
