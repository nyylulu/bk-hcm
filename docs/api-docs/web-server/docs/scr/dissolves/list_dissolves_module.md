### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：机房裁撤模块。

### URL

GET /api/v1/woa/dissolves/module

### 输入参数

无

### 调用示例

无

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
        "module_name": "广州-亚太-M1,美国-加州-M2,广州-亚太-M3,广州-人民中-M1",
        "start_time": "2018-11-01",
        "end_time": "2019-03-20",
        "which_stages": 1
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
| items       | object array | 裁撤模块列表  |

#### data.items

| 参数名称       | 参数类型   | 描述         |
|---------------|----------|--------------|
| module_name	| string   | 机房模块信息   |
| start_time	| string   | 开始时间      |
| end_time	    | string   | 结束时间      |
| which_stages	| int	   | 排序          |
