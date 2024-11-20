### 描述

- 该接口提供版本：v1.6.11+。
- 该接口所需权限：平台管理-主机回收。
- 该接口功能描述：按指定项目类型进行资源回收。

### URL

POST /api/v1/woa/task/start/recycle/order/by/recycle_type

### 输入参数

| 参数名称              | 参数类型         | 必选 | 描述                   |
|-------------------|--------------|----|----------------------|
| suborder_id_types | object array | 是  | 回收子单据ID跟数组，数量最大限制100 |

#### sub_order_id_types

| 参数名称         | 参数类型   | 必选 | 描述                                 |
|--------------|--------|----|------------------------------------|
| suborder_id  | string | 是  | 回收子单据ID                            |
| recycle_type | string | 是  | 回收类型(枚举值:常规项目、机房裁撤、过保裁撤、春节保障、滚服项目) |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "suborder_id_types": [
    {
      "suborder_id": "1001-1",
      "recycle_type": "常规项目"
    }
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": null
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 请求返回的数据                    |

#### data

无
