### 描述

- 该接口提供版本：v1.7.4+。
- 该接口所需权限：平台-资源预测。
- 该接口功能描述：根据机型名称从CRP同步预测的机型。

### URL

POST /api/v1/woa/plans/device_types/sync

### 输入参数

| 参数名称         | 参数类型         | 必选 | 描述           |
|--------------|--------------|----|--------------|
| device_types | string array | 是  | 机型名称，数量最大100 |

### 调用示例

```json
{
  "device_types": [
    "S5.2XLARGE16",
    "S5.8XLARGE32"
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述                        |
|---------|--------|---------------------------|
| code    | int    | 错误编码。 0表示success，>0表示失败错误 |
| message | string | 请求失败返回的错误信息               |
| data	   | object | 响应数据                      |

#### data

无
