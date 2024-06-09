### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：机房裁撤。
- 该接口功能描述：查询整体裁撤表格信息。

### URL

POST /api/v1/woa/dissolve/table/list

### 输入参数

| 参数名称         | 参数类型         | 必选 | 描述     |
|--------------|--------------|----|--------|
| organizations      | string array | 否  | 组织架构，以“/”分割，至少包含三级，最多包含五级，如："A公司/B事业群/C部门/D中心/F组" |
| bk_biz_names | string array | 否  | 业务名称   |
| module_names | string array | 是  | 裁撤模块名称 |
| operators    | string array | 否  | 人员名称   |

### 调用示例

查询组织架构路径为“A公司/B事业群/C部门/D中心/F组”, 业务名称为biz, 裁撤模块名称为module, operator为test的整体裁撤表格信息。

```json
{
  "organizations": ["A公司/B事业群/C部门/D中心/F组"],
  "bk_biz_names": ["biz"],
  "module_names": ["module"],
  "operators": ["test"]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "",
  "data": {
    "items": [
      {
        "bk_biz_name": "biz",
        "module_host_count": {
          "module": 8
        },
        "total": {
          "origin": {
            "host_count": 8,
            "cpu_count": 640
          },
          "current": {
            "host_count": 0,
            "cpu_count": 0
          }
        },
        "progress": "100.00%"
      },
      {
        "bk_biz_name": "总数",
        "module_host_count": {
          "module": 8
        },
        "total": {
          "origin": {
            "host_count": 8,
            "cpu_count": 640
          },
          "current": {
            "host_count": 0,
            "cpu_count": 0
          }
        },
        "progress": ""
      },
      {
        "bk_biz_name": "裁撤进度",
        "module_list": {},
        "total": {
          "origin": {
            "host_count": "100.00%",
            "cpu_count": 0
          },
          "current": {
            "host_count": "100.00%",
            "cpu_count": 0
          }
        },
        "progress": "100.00%"
      }
    ]
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

| 参数名称  | 参数类型   | 描述         |
|-------|--------|------------|
| items | array  | 业务裁撤进度相关数据 |

#### data.items[n]

| 参数名称              | 参数类型           | 描述                       |
|-------------------|----------------|--------------------------|
| bk_biz_name       | string         | 业务名称                     |
| module_host_count | map[string]int | key为模块名称，value为该模块下的主机数量 |
| total             | object         | 主机总数相关信息                 |
| progress | string         | 裁撤进度 |

#### data.items[n].total

| 参数名称  | 参数类型   | 描述   |
|---------|--------|-------|
| current | object | 当前   |
| origin  | object    | 原始   |

#### data.items[n].total.current

| 参数名称  | 参数类型          | 描述                                                         |
|---------|---------------|------------------------------------------------------------|
| host_count | int or string | 当bk_biz_name为"裁撤进度"时，该字段为string类型，表示主机裁撤进度；否则为int类型，表示主机数量 |
| cpu_count  | int           | cpu核心数                                                     |

#### data.items[n].total.origin

| 参数名称  | 参数类型          | 描述                                                         |
|---------|---------------|------------------------------------------------------------|
| host_count | int or string | 当bk_biz_name为"裁撤进度"时，该字段为string类型，表示主机裁撤进度；否则为int类型，表示主机数量 |
| cpu_count  | int           | cpu核心数                                                     |