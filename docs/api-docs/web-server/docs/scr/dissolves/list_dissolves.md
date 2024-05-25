### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：裁撤列表。

### URL

POST /api/v1/woa/dissolves

### 输入参数

| 参数名称     | 参数类型       | 必选 | 描述        |
|-------------|--------------|------|------------|
| group_name  | string       | 是   | 业务组      |
| module_list | string array | 是   | 模块列表     |
| offset	  | int	         | 否   | 偏移量       |
| limit	      | int	         | 否   | 返回条目个数  |

### 调用示例

#### 获取详细信息请求参数示例

```json
{
  "group_name": "蓝鲸运营组",
  "module_list": ["广州-人民中-M1","美国-加州-M2","广州-亚太-M3","广州-亚太-M1"]
}
```

### 响应示例

#### 获取详细信息返回结果示例

```json
{
  "code":0,
  "message":"OK",
  "data":{
    "total_count": 12,
    "items": [
      {
        "display_name": "GSE",
        "module_list": [
          {
            "module_name": "广州-亚太-M1",
            "count": 1
          },
          {
            "module_name": "广州-亚太-M3",
            "count": 1
          },
          {
            "module_name": "广州-人民中-M1",
            "count": 0
          },
          {
            "module_name": "美国-加州-M2",
            "count": 0
          }
        ],
        "total": {
          "origin": 4,
          "current": 2
        },
        "delived": 0,
        "progress": "50.00%"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称    | 参数类型       | 描述               |
|------------|--------------|--------------------|
| code       | int          | 错误编码。 0表示success，>0表示失败错误  |
| message    | string       | 请求失败返回的错误信息 |
| data	     | object array | 响应数据             |

#### data

| 参数名称     | 参数类型       | 描述    |
|-------------|--------------|---------|
| total_count | int          | 总数     |
| items       | object array | 裁撤列表  |

#### data.items

| 参数名称      | 参数类型       | 描述      |
|--------------|--------------|-----------|
| display_name | string       | 业务名称   |
| module_list  | object array | 模块列表   |
| total	       | object	      | 总数      |
| delived	   | int          | 已交付设备 |
| progress	   | string       | 裁撤进度   |

#### data.items.module_list

| 参数名称     | 参数类型  | 描述    |
|-------------|---------|---------|
| module_name | string  | 模块名称 |
| count	      | int     | 个数    |

#### data.items.total

| 参数名称  | 参数类型 | 描述   |
|---------|---------|-------|
| current | int     | 当前   |
| origin  | int     | 原始   |
