### 描述

- 该接口提供版本：v1.8.5.6+。
- 该接口所需权限：平台-资源预测。
- 该接口功能描述：查询资源下转移额度使用概览信息。

### URL

POST /api/v1/woa/plans/resources/transfer_quotas/summary

### 输入参数

| 参数名称           | 参数类型        | 必选  | 描述                             |
|-------------------|---------------|-------|---------------------------------|
| bk_biz_id         | int array     | 否    | 业务ID，最多100个                 |
| year              | int           | 是    | 额度所属年份                      |
| applied_type      | string array  | 否    | 转移类型（枚举值：add(转移进池)、remove(转移出池)）  |
| sub_ticket_id     | string array  | 否    | 预测调整子单号                    |
| technical_class   | string array  | 否    | 技术分类                         |
| obs_project       | string array  | 否    | 项目类型                         |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "year": 2025,
  "technical_class": [
    "标准型"
  ],
  "obs_project": [
    "常规项目"
  ]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "ok",
  "data": {
      "used_quota": 100
      "remain_quota": 100
  }
}
```

### 响应参数说明

| 参数名称  | 参数类型   | 描述   |
|---------|-----------|--------|
| code    | int       | 状态码  |
| message | string    | 请求信息 |
| data    | object    | 响应数据 |

#### data

| 参数名称      | 参数类型   | 描述        |
|--------------|----------|-------------|
| used_quota   | int      | 已使用额度   |
| remain_quota | int      | 剩余额度     |
