### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：CVM资源最大可申请量查询。

### 输入参数
| 字段          | 类型   | 必选   |  描述                                      |
|--------------|--------|-------|--------------------------------------------|
| require_type | int    | 是    | 需求类型。1: 常规项目; 2: 春节保障; 3: 机房裁撤  |
| region       | string | 是    | 地域                                        |
| zone         | string | 是    | 可用区                                      |
| device_type  | string | 是    | 机型                                        |
| vpc          | string | 否    | vpc。若vpc为空，则返回IEG默认vpc的最大可申领量   |
| subnet       | string | 否    | 子网。若vpc不为空且subnet为空，则返回vpc下所有子网的最大可申领量 |

### 请求示例
```json
{
  "require_type":1,
  "region":"ap-shanghai",
  "zone":"ap-shanghai-2",
  "device_type":"S3ne.4XLARGE64",
  "vpc":"",
  "subnet":""
}
```

### 响应示例
```json
{
  "result":true,
  "code":0,
  "message":"success",
  "permission":null,
  "request_id":"f5a6331d4bc2433587a63390c76ba7bf",
  "data":{
    "info":[
        {
            "region":"ap-shanghai",
            "zone":"ap-shanghai-2",
            "vpc":"",
            "subnet":"",
            "max_num":45,
            "max_info":[
              {
                "key":"云后端CBS容量计算可申领量",
                "value":720
              },
              {
                "key":"云后端CVM容量计算可申领量",
                "value":720
              },
              {
                "key":"所选VPC子网可用IP数",
                "value":45
              },
              {
                "key":"未执行需求预测的可申领量",
                "value":720
              },
              {
                "key":"云梯系统单次提单最大量",
                "value":1000
              }
          ]
      }
    ]
  }
}
```

### 响应参数说明
| 参数名称     | 参数类型 | 描述                                 |
| ------------| -------| -------------------------------------|
| result      | bool   | 请求成功与否。true:请求成功；false请求失败 |
| code        | int    | 错误编码。 0表示success，>0表示失败错误   |
| message     | string | 请求失败返回的错误信息                   |
| permission  | object | 权限信息                               |
| request_id  | string | 请求链id                               |
| data        | object | 请求返回的数据                          |

#### data 字段说明：

| 名称  | 类型         | 说明                      |
|------|--------------|--------------------------|
| info | object array | 资源最大可申请量信息详情列表  |

#### info 字段说明：
| 名称      | 类型          | 说明             |
|----------|--------------|------------------|
| region   | string       | 城市              |
| zone     | string       | 可用区             |
| vpc      | string       | vpc。若vpc为空，则返回可用区下所有vpc的最大可申领量 |
| subnet   | string       | 子网。若vpc不为空且subnet为空，则返回vpc下所有子网的最大可申领量 |
| max_num  | int          | 资源最大可申请量     |
| max_info | object array | 资源最大可申请量详情  |

#### max_info 字段说明：
| 名称   | 类型   | 说明                    |
|-------|--------|------------------------|
| key   | string | 资源最大可申请量的维度的键 |
| value | int    | 资源最大可申请量的维度的值 |
