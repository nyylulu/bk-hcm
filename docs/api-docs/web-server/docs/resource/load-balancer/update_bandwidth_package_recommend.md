### 描述

- 该接口提供版本：v1.8.5.9+。
- 该接口所需权限：平台-全局配置。
- 该接口功能描述：更新带宽包推荐配置, 覆盖式更新。

### URL

PATCH /api/v1/cloud/bandwidth_packages/recommend

### 输入参数


| 参数名称        | 参数类型         | 必选 | 描述                 |
|-------------|--------------|----|--------------------|
| package_ids | string array | 是  | 推荐的带宽包id, 最长限制100个 |

### 调用示例

#### tcloud

```json
{
  "package_ids": ["bwp-xxxxxx"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |
