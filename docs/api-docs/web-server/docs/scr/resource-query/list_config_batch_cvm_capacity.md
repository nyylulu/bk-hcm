### 描述

- 该接口提供版本：v1.8.7.0+。
- 该接口所需权限：无。
- 该接口功能描述：支持多机型、多可用区的并发库存查询。

### URL

POST /api/v1/woa/config/findmany/cvm/capacity

### 输入参数

| 参数名称         | 参数类型         | 必选 | 描述                                                                          |
|--------------|--------------|----|-----------------------------------------------------------------------------|
| require_type | int	         | 是	 | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤; 6: 滚服项目; 7: 小额绿通; 8: 春保资源池                  |
| device_types | array string | 是	 | 设备类型列表，支持多个机型同时查询。长度限制：1-100个设备类型                                           |
| region	      | string	      | 是	 | 地域                                                                          |
| zones	       | array string | 是	 | 多可用区列表。支持同时查询多个可用区的容量信息。,长度限制：1-100个可用区。注意：当前版本不支持不传此参数查询全部可用区，必须指定具体的可用区列表 |
| vpc	         | string	      | 否	 | vpc。若vpc为空，则返回IEG默认vpc的最大可申领量                                               |
| subnet	      | string	      | 否	 | 子网。若vpc不为空且subnet为空，则返回vpc下所有子网的最大可申领量                                      |
| charge_type  | string       | 否  | 计费模式 (PREPAID:包年包月，POSTPAID_BY_HOUR:按量计费)，默认:包年包月                           |

### 调用示例

#### 单机型、多可用区查询请求参数示例

```json
{
  "require_type": 1,
  "device_types": ["D4-15-400-10"],
  "region": "ap-guangzhou",
  "zones": ["ap-guangzhou-1", "ap-guangzhou-2"],
  "vpc": "vpc-12345678",
  "subnet": "subnet-12345678",
  "charge_type": "PREPAID"
}
```

#### 多机型、多可用区查询请求参数示例

```json
{
  "require_type": 1,
  "device_types": ["D4-15-400-10", "D8-8-500-10"],
  "region": "ap-guangzhou",
  "zones": ["ap-guangzhou-1", "ap-guangzhou-2"],
  "vpc": "vpc-12345678",
  "subnet": "subnet-12345678",
  "charge_type": "PREPAID"
}
```

### 响应示例

#### 单机型、多可用区查询返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "count": 3,
    "info": [
      {
        "device_type": "D4-15-400-10",
        "region": "ap-guangzhou",
        "zone": "ap-guangzhou-1",
        "vpc": "vpc-12345678",
        "subnet": "subnet-12345678",
        "max_num": 100,
        "max_info": [
          {
            "key": "云后端CBS容量计算可申领量",
            "value": 100
          },
          {
            "key": "云后端CVM容量计算可申领量",
            "value": 100
          },
          {
            "key": "所选VPC子网可用IP数",
            "value": 50
          },
          {
            "key": "未执行需求预测的可申领量",
            "value": 100
          },
          {
            "key": "云梯系统单次提单最大量",
            "value": 1000
          }
        ]
      },
      {
        "device_type": "D4-15-400-10",
        "region": "ap-guangzhou",
        "zone": "ap-guangzhou-2",
        "vpc": "vpc-12345678",
        "subnet": "subnet-12345678",
        "max_num": 80,
        "max_info": [
          {
            "key": "云后端CBS容量计算可申领量",
            "value": 80
          },
          {
            "key": "云后端CVM容量计算可申领量",
            "value": 80
          },
          {
            "key": "所选VPC子网可用IP数",
            "value": 60
          },
          {
            "key": "未执行需求预测的可申领量",
            "value": 80
          },
          {
            "key": "云梯系统单次提单最大量",
            "value": 1000
          }
        ]
      }
    ]
  }
}
```

#### 多机型、多可用区查询返回结果示例

```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "count": 4,
    "info": [
      {
        "device_type": "D4-15-400-10",
        "region": "ap-guangzhou",
        "zone": "ap-guangzhou-1",
        "vpc": "vpc-12345678",
        "subnet": "subnet-12345678",
        "max_num": 100,
        "max_info": [
          {
            "key": "云后端CBS容量计算可申领量",
            "value": 100
          },
          {
            "key": "云后端CVM容量计算可申领量",
            "value": 100
          },
          {
            "key": "所选VPC子网可用IP数",
            "value": 50
          },
          {
            "key": "未执行需求预测的可申领量",
            "value": 100
          },
          {
            "key": "云梯系统单次提单最大量",
            "value": 1000
          }
        ]
      },
      {
        "device_type": "D4-15-400-10",
        "region": "ap-guangzhou",
        "zone": "ap-guangzhou-2",
        "vpc": "vpc-12345678",
        "subnet": "subnet-12345678",
        "max_num": 80,
        "max_info": [
          {
            "key": "云后端CBS容量计算可申领量",
            "value": 80
          },
          {
            "key": "云后端CVM容量计算可申领量",
            "value": 80
          },
          {
            "key": "所选VPC子网可用IP数",
            "value": 60
          },
          {
            "key": "未执行需求预测的可申领量",
            "value": 80
          },
          {
            "key": "云梯系统单次提单最大量",
            "value": 1000
          }
        ]
      },
      {
        "device_type": "D8-8-500-10",
        "region": "ap-guangzhou",
        "zone": "ap-guangzhou-1",
        "vpc": "vpc-12345678",
        "subnet": "subnet-12345678",
        "max_num": 200,
        "max_info": [
          {
            "key": "云后端CBS容量计算可申领量",
            "value": 200
          },
          {
            "key": "云后端CVM容量计算可申领量",
            "value": 200
          },
          {
            "key": "所选VPC子网可用IP数",
            "value": 50
          },
          {
            "key": "未执行需求预测的可申领量",
            "value": 200
          },
          {
            "key": "云梯系统单次提单最大量",
            "value": 1000
          }
        ]
      },
      {
        "device_type": "D8-8-500-10",
        "region": "ap-guangzhou",
        "zone": "ap-guangzhou-2",
        "vpc": "vpc-12345678",
        "subnet": "subnet-12345678",
        "max_num": 150,
        "max_info": [
          {
            "key": "云后端CBS容量计算可申领量",
            "value": 150
          },
          {
            "key": "云后端CVM容量计算可申领量",
            "value": 150
          },
          {
            "key": "所选VPC子网可用IP数",
            "value": 60
          },
          {
            "key": "未执行需求预测的可申领量",
            "value": 150
          },
          {
            "key": "云梯系统单次提单最大量",
            "value": 1000
          }
        ]
      }
    ]
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

| 参数名称  | 参数类型         | 描述             |
|-------|--------------|----------------|
| count | int          | 当前规则能匹配到的总记录条数 |
| info  | object array | 资源最大可申请量信息详情列表 |

#### data.info (批量查询)

| 参数名称        | 参数类型         | 描述                                     |
|-------------|--------------|----------------------------------------|
| device_type | string	      | 设备类型                                   |
| region	     | string	      | 城市                                     |
| zone	       | string	      | 可用区                                    |
| vpc	        | string	      | vpc。若vpc为空，则返回可用区下所有vpc的最大可申领量         |
| subnet	     | string	      | 子网。若vpc不为空且subnet为空，则返回vpc下所有子网的最大可申领量 |
| max_num	    | int	         | 资源最大可申请量                               |
| max_info	   | object array | 资源最大可申请量详情                             |