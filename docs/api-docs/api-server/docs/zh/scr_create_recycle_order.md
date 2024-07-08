### 描述

- 该接口提供版本：v1.6.0+。
- 该接口所需权限：业务-主机回收。
- 该接口功能描述：创建资源回收单据。

### URL

POST /api/v1/woa/task/create/recycle/order

### 输入参数

| 参数名称       | 参数类型       | 必选 | 描述             |
|---------------|--------------|------|-----------------|
| bk_biz_id     | int	       | 是	  | CC业务ID         |
| ips           | string array | 否	  | 回收资源ip列表     |
| asset_ids	    | string array | 否	  | 回收资源固资号列表  |
| bk_host_ids	| int array	   | 否	  | 回收资源CC主机ID   |
| remark	    | string       | 否	  | 备注              |

- 说明：ips、asset_ids和bk_host_ids不能同时为空

### 调用示例

#### 获取详细信息请求参数示例

```json
{
"bk_biz_id": 3,
"ips": [
"10.0.0.1"
],
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
"permission":null,
"request_id":"f5a6331d4bc2433587a63390c76ba7bf",
"data":{
"order_id": 1001
}
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| result     | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| permission | object       | 权限信息             |
| request_id | string       | 请求链ID             |
| data	     | object array | 响应数据             |

#### data

| 参数名称  | 参数类型 | 描述   |
|----------|--------|--------|
| order_id | int    | 单据ID |
