### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务-IaaS资源创建。
- 该接口功能描述：校验预测并创建CRP升降配单据。

### URL

POST /api/v1/woa/bizs/{bk_biz_id}/task/create/upgrade/crp_order

### 输入参数

| 参数名称              | 参数类型         | 必选 | 描述              |
|-------------------|--------------|----|-----------------|
| bk_biz_id         | int64        | 是  | 业务ID            |
| remark	           | string       | 否	 | 备注              |
| upgrade_cvm_list	 | object array | 是  | 升降配目标列表，数量最大500 |

#### upgrade_cvm_list[i]

##### bk_host_id 和 instance_id 必须且只能提供其中一个

| 参数名称                 | 参数类型    | 必选 | 描述             |
|----------------------|---------|----|----------------|
| bk_host_id           | int64	  | 否  | 升降配实例对应的CC主机ID |
| instance_id          | string	 | 否  | 升降配实例的云实例ID    |
| target_instance_type | string	 | 是  | 调整的目标实例类型      |

### 调用示例

```json
{
  "remark": "xx",
  "upgrade_cvm_list": [
    {
      "instance_id": "ins-xxxxx",
      "target_instance_type": "SA5.8XLARGE96"
    },
    {
      "bk_host_id": 1111111,
      "target_instance_type": "SA5t.2XLARGE16"
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
  "data": {
    "crp_order_id": "TZ2025060511560xxxxxx"
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型         | 描述                         |
|---------|--------------|----------------------------|
| result  | bool         | 请求成功与否。true:请求成功；false请求失败 |
| code    | int          | 错误编码。 0表示success，>0表示失败错误  |
| message | string       | 请求失败返回的错误信息                |
| data	   | object array | 响应数据                       |

#### data

| 参数名称         | 参数类型   | 描述      |
|--------------|--------|---------|
| crp_order_id | string | CRP单据ID |
