### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：云账单拉取。
- 该接口功能描述：查询AI相关云账单接口列表。

### 输入参数

#### url 参数

| 参数名称   | 参数类型   | 必选 | 描述  |
|--------|--------|----|-----|
| vendor | string | 是  | 云厂商 |

##### vendor 列表：

- aws
- gcp

#### Body参数

| 参数名称           | 参数类型   | 必选 | 描述                |
|----------------|--------|----|-------------------|
| bill_year      | uint   | 是  | 账单年份              |
| bill_month     | uint   | 是  | 账单月份              |
| begin_bill_day | uint   | 否  | 账单开始日，需和账单截止日一起设定 |
| end_bill_day   | uint   | 否  | 账单截止日，需和账单开始日一起设定 |
| page           | object | 是  | 分页设置              |

##### page

| 参数名称  | 参数类型   | 必选 | 描述                                                                                                                                                  |
|-------|--------|----|-----------------------------------------------------------------------------------------------------------------------------------------------------|
| count | bool   | 是  | 是否返回总记录条数。 如果为true，查询结果返回总记录条数 count，但查询结果详情数据 details 为空数组，此时 start 和 limit 参数将无效，且必需设置为0。如果为false，则根据 start 和 limit 参数，返回查询结果详情数据，但总记录条数 count 为0 |
| start | uint32 | 否  | 记录开始位置，start 起始值为0                                                                                                                                  |
| limit | uint32 | 否  | 每页限制条数，最大500，不能为0                                                                                                                                   |
| sort  | string | 否  | 排序字段，返回数据将按该字段进行排序                                                                                                                                  |
| order | string | 否  | 排序顺序（枚举值：ASC、DESC）                                                                                                                                  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "bill_year": 2024,
  "bill_month": 7,
  "begin_bill_day": 30,
  "end_bill_day": 30,
  "page": {
    "limit": 100,
    "start": 0,
    "count": false
  }
}

```

### 响应参数说明

| 参数名称    | 参数类型   | 描述   |
|---------|--------|------|
| code    | int32  | 状态码  |
| message | string | 请求信息 |
| data    | object | 响应数据 |

#### data

| 参数名称   | 参数类型   | 描述             |
|--------|--------|----------------|
| count  | uint64 | 当前规则能匹配到的总记录条数 |
| detail | array  | 查询返回的数据        |

#### detail

| 参数名称                  | 参数类型   | 描述                             |
|-----------------------|--------|--------------------------------|
| id                    | string | 账单id                           |
| vendor                | string | 云厂商                            |
| product_id            | int32  | 产品id                           |
| year                  | int32  | 账单年份, 如: 2024                  |
| month                 | int32  | 账单月份, 如: 7                     |
| day                   | int32  | 账单日, 如: 1                      |
| main_account_cloud_id | string | 二级账号ID                         |
| main_account_email    | string | 二级账号邮箱地址                       |
| main_account_name     | string | 二级账号名                          |
| llm_type              | string | Claude/Gemini/Unknown          |
| currency              | string | 货币类型, 如USD                     |
| cost                  | string | 原货币金额                          |
| rate                  | string | 汇率                             |
| cost_rmb              | string | 按汇率转将原货币换为人民币的金额               |
| updated_at            | string | 更新时间，标准格式：2006-01-02T15:04:05Z |
| raw_bill              | object | 云上原始账单格式，见下文                   |

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "count": 1,
    "details": [
      {
        "id": "0001234f",
        "year": 2025,
        "month": 2,
        "day": 3,
        "vendor": "aws",
        "product_id": 12345,
        "main_account_cloud_id": "01234567891234",
        "main_account_email": "xxx@xxx.com",
        "main_account_name": "xxx",
        "llm_type": "Claude",
        "cost": "0.1",
        "rate": "7.0",
        "cost_rmb": "0.7",
        "currency": "USD",
        "updated_at": "2025-03-04T12:19:50Z",
        "raw_bill": {
          "pricing_term": "OnDemand",
          "pricing_unit": "Units",
          "product_region": "us-west-2",
          "bill_invoice_id": "1234576899",
          "product_location": "",
          "bill_billing_entity": "AWS Marketplace",
          "line_item_operation": "Usage",
          "line_item_usage_type": "USW2-MP:USW2_InputTokenCount-Units",
          "product_product_name": "Claude 3.5 Sonnet (Amazon Bedrock Edition)",
          "bill_payer_account_id": "1234567890",
          "identity_line_item_id": "xxxxxxxxxxxxxxxxx",
          "line_item_resource_id": "arn:aws:bedrock:us-west-2::foundation-model/anthropic.claude-3-5-sonnet-20240620-v1:0",
          "product_instance_type": "",
          "line_item_product_code": "4a524th30q694o538qndh9ucht",
          "line_item_usage_amount": "0.02645",
          "product_product_family": "",
          "line_item_currency_code": "USD",
          "line_item_line_item_type": "Usage",
          "line_item_unblended_cost": "0.1",
          "line_item_unblended_rate": "1.0000000000",
          "line_item_usage_end_date": "",
          "line_item_usage_account_id": "01234567891234",
          "line_item_usage_start_date": "",
          "reservation_effective_cost": "0.0",
          "line_item_net_unblended_cost": "0.1",
          "line_item_net_unblended_rate": "1.0000000000",
          "pricing_public_on_demand_cost": "0.0",
          "pricing_public_on_demand_rate": "",
          "reservation_net_effective_cost": "0.0",
          "savings_plan_savings_plan_rate": "0.0",
          "line_item_line_item_description": "AWS Marketplace software usage|us-west-2|Million Input Tokens",
          "savings_plan_savings_plan_a_r_n": "",
          "savings_plan_savings_plan_effective_cost": "0.0",
          "savings_plan_net_savings_plan_effective_cost": "0.0"
        }
      }
    ]
  }
}
```

#### AWS raw_bill 说明

| 参数名称                                                             | 参数类型   | 描述                  |
|------------------------------------------------------------------|--------|---------------------|
| bill_bill_type                                                   | string | 计费类别                |
| bill_billing_entity                                              | string | 账单实体                |
| bill_billing_period_end_date                                     | string | 账单周期截止日期            |
| bill_billing_period_start_date                                   | string | 账单周期开始日期            |
| bill_invoice_id                                                  | string | 账单清单ID              |
| bill_invoicing_entity                                            | string | 账单清单实体              |
| bill_payer_account_id                                            | string | 账单支付账号ID            |
| discount_edp_discount                                            | string | edp优惠金额             |
| discount_total_discount                                          | string | 总优惠金额               |
| identity_line_item_id                                            | string | 项目ID                |
| identity_time_interval                                           | string | 标识时间间隔              |
| line_item_availability_zone                                      | string | 可用区                 |
| line_item_blended_cost                                           | string | 混合成本                |
| line_item_blended_rate                                           | string | 混合费率                |
| line_item_currency_code                                          | int    | 项目当前代码              |
| line_item_legal_entity                                           | string | 项目合法实体              |
| line_item_line_item_description                                  | string | 计费描述                |
| line_item_line_item_type                                         | string | 项目类型                |
| line_item_net_unblended_cost                                     | string | 未混合折后成本             |
| line_item_net_unblended_rate                                     | string | 未混合折后费率             |
| line_item_normalization_factor                                   | string | 项目规范因子              |
| line_item_normalized_usage_amount                                | string | 规范化使用费用             |
| line_item_operation                                              | string | 项目操作                |
| line_item_product_code                                           | string | 项目产品代码              |
| line_item_resource_id                                            | string | 资源ID                |
| line_item_tax_type                                               | string | 项目税费类型              |
| line_item_unblended_cost                                         | string | 未混合成本               |
| line_item_unblended_rate                                         | string | 未混合费率               |
| line_item_usage_account_id                                       | string | 使用的账号ID             |
| line_item_usage_amount                                           | string | 使用金额                |
| line_item_usage_end_date                                         | string | 使用截止日期              |
| line_item_usage_start_date                                       | string | 使用开始日期              |
| line_item_usage_type                                             | string | 使用类型                |
| pricing_currency                                                 | string | 定价货币                |
| pricing_lease_contract_length                                    | string | 定价合同长度              |
| pricing_offering_class                                           | string | 报价类别                |
| pricing_public_on_demand_cost                                    | string | 定价公开需求成本            |
| pricing_public_on_demand_rate                                    | string | 定价公开需求费率            |
| pricing_purchase_option                                          | string | 付款方式：全量预付、部分预付、无预付  |
| pricing_term                                                     | string | 使用量是预留还是按需          |
| pricing_unit                                                     | string | 费用计价单位              |
| product_database_engine                                          | string | 产品数据库引擎             |
| product_from_location                                            | string | 产品来源                |
| product_from_location_type                                       | string | 产品来源类型              |
| product_from_region_code                                         | string | 产品区域编码              |
| product_instance_type                                            | string | 产品实例类型              |
| product_instance_type_family                                     | string | 产品实例类型系列            |
| product_location                                                 | string | 产品定位                |
| product_location_type                                            | string | 产品定位类型              |
| product_marketoption                                             | string | 市场选项                |
| product_normalization_size_factor                                | string | 产品规格因子              |
| product_operation                                                | string | 产品操作                |
| product_product_family                                           | string | 产品系列                |
| product_product_name                                             | string | 产品名称                |
| product_purchase_option                                          | string | 产品采购选项              |
| product_purchaseterm                                             | string | 产品采购条款              |
| product_region                                                   | string | 产品区域                |
| product_region_code                                              | string | 产品区域编码              |
| product_servicecode                                              | string | 产品服务编码              |
| product_servicename                                              | string | 产品服务名称              |
| product_tenancy                                                  | string | 产品库存                |
| product_to_location                                              | string | 产品指向的位置             |
| product_to_location_type                                         | string | 产品指向的位置类型           |
| product_to_region_code                                           | string | 产品指向的区域的编码          |
| product_transfer_type                                            | string | 产品传输类型              |
| reservation_amortized_upfront_cost_for_usage                     | string | 预留摊销前期使用成本          |
| reservation_amortized_upfront_fee_for_billing_period             | string | 预留摊销预付费账单周期         |
| reservation_effective_cost                                       | string | 预留有效成本              |
| reservation_end_time                                             | string | 预留截止时间              |
| reservation_modification_status                                  | string | 预留修改状态              |
| reservation_net_amortized_upfront_cost_for_usage                 | string | 预留网络摊销可用成本          |
| reservation_net_amortized_upfront_fee_for_billing_period         | string | 预留网络摊销预付费账单周期       |
| reservation_net_effective_cost                                   | string | 预留网络有效成本            |
| reservation_net_recurring_fee_for_usage                          | string | 预留可用的常用费用           |
| reservation_net_unused_amortized_upfront_fee_for_billing_period  | string | 预留网络未使用预付费账单周期      |
| reservation_net_unused_recurring_fee                             | string | 预留网络未使用常用费用         |
| reservation_net_upfront_value                                    | string | 预留前期净值              |
| reservation_normalized_units_per_reservation                     | string | 预留规范化单位每次保留量        |
| reservation_number_of_reservations                               | string | 预留数量                |
| reservation_recurring_fee_for_usage                              | string | 预留可用的常用费用           |
| reservation_reservation_a_r_n                                    | string | 预留的ARN              |
| reservation_start_time                                           | string | 预留开始时间              |
| reservation_subscription_id                                      | string | 预留的订阅ID             |
| reservation_total_reserved_normalized_units                      | string | 预留总服务标准化单位          |
| reservation_total_reserved_units                                 | string | 预留总服务单位             |
| reservation_units_per_reservation                                | string | 预留每次保留的单位           |
| reservation_unused_amortized_upfront_fee_for_billing_period      | string | 预留未使用冻结的预付费计费周期     |
| reservation_unused_normalized_unit_quantity                      | string | 预留未使用规范化单位数量        |
| reservation_unused_quantity                                      | string | 预留未使用的数量            |
| reservation_unused_recurring_fee                                 | string | 预留未使用的现金            |
| reservation_upfront_value                                        | string | 预留上行数值              |
| savings_plan_amortized_upfront_commitment_for_billing_period     | string | 账单期的计划摊销前期承诺        |
| savings_plan_end_time                                            | string | SavingsPlan截止时间     |
| savings_plan_net_amortized_upfront_commitment_for_billing_period | string | SavingsPlan承诺账单周期   |
| savings_plan_net_recurring_commitment_for_billing_period         | string | SavingsPlan计划净现金周期  |
| savings_plan_net_savings_plan_effective_cost                     | string | SavingsPlan网络有效成本   |
| savings_plan_offering_type                                       | string | SavingsPlan报价类型     |
| savings_plan_payment_option                                      | string | SavingsPlan支付类型     |
| savings_plan_purchase_term                                       | string | SavingsPlan采购期限     |
| savings_plan_recurring_commitment_for_billing_period             | string | SavingsPlan现金承诺账单周期 |
| savings_plan_region                                              | string | SavingsPlan区域       |
| savings_plan_savings_plan_a_r_n                                  | string | SavingsPlanARN      |
| savings_plan_savings_plan_effective_cost                         | string | SavingsPlan有效成本     |
| savings_plan_savings_plan_rate                                   | string | SavingsPlan计划费率     |
| savings_plan_start_time                                          | string | SavingsPlan计划开始时间   |
| savings_plan_total_commitment_to_date                            | string | SavingsPlan总承诺日期    |
| savings_plan_used_commitment                                     | string | SavingsPlan已使用承诺    |


#### GCP raw_bill 说明

| 参数名称                          | 参数类型    | 描述                             |
|-------------------------------|---------|--------------------------------|
| billing_account_id            | string  | 与使用量相关的 Cloud Billing 帐号ID     |
| cost                          | float64 | 成本                             |
| cost_type                     | string  | 费用类型                           |
| country                       | string  | 国家                             |
| credits                       | json    | 赠送金信息                          |
| currency                      | string  | 币种                             |
| location                      | string  | 区域信息                           |
| month                         | string  | 账单年月                           |
| project_id                    | string  | 项目ID                           |
| project_name                  | string  | 项目名称                           |
| project_number                | string  | 项目编号                           |
| region                        | string  | 区域                             |
| resource_global_name          | string  | 资源全局唯一标识符                      |
| resource_name                 | string  | 资源名称                           |
| service_description           | string  | 服务描述                           |
| service_id                    | string  | 服务ID                           |
| sku_description               | string  | 资源类型描述                         |
| sku_id                        | string  | 资源类型ID                         |
| usage_amount                  | string  | 可用金额                           |
| usage_amount_in_pricing_units | string  | 可用金额单价                         |
| usage_end_time                | string  | 可用结束时间，示例：2023-04-16T15:00:00Z |
| usage_pricing_unit            | float64 | 可用金额单价的单位                      |
| usage_start_time              | string  | 可用开始时间，示例：2023-04-16T15:00:00Z |
| usage_unit                    | string  | 可用金额单位                         |
| zone                          | string  | 可用区                            |