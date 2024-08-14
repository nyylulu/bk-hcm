### 描述

- 该接口提供版本：v1.6.1+。
- 该接口所需权限：平台-CVM机型。
- 该接口功能描述：CVM机型配置信息更新。

### URL

PUT /api/v1/woa/config/updatemany/config/cvm/device/property

### 输入参数

| 参数名称     | 参数类型                | 必选 | 描述          |
|-------------|-----------------------|------|--------------|
| ids         | int array             | 是   | 更新的配置实例ID列表，最大为20  |
| properties  |map[string]interface{} | 是   | 更新的属性      |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "ids":[
    1,
    2,
    3
  ],
  "properties":{
    "enable_capacity":false,
    "enable_apply":false,
    "score":90.5,
    "comment":"disable reason"
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
