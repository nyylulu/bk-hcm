### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询用户在当前地域负载均衡集群中的资源列表。腾讯云代理接口 DescribeClusterResources

### URL

POST /api/v1/cloud/vendors/tcloud-ziyan/load_balancers/cluster_resources/describe

### 输入参数

| 参数名称             | 参数类型         | 必选 | 描述                       |
|------------------|--------------|----|--------------------------|
| account_id       | string       | 是  | 云账户id                    |
| region           | string       | 是  | 地域                       |
| cluster_id       | string array | 否  | 按照 集群 的唯一ID过滤            |
| vip              | string array | 否  | 按照vip过滤                  |
| load_balancer_id | string array | 否  | 按照负载均衡唯一ID过滤             |
| idle             | bool         | 否  | 按照是否闲置过滤，如"true","false" |
| limit            | int          | 否  | 返回资源列表数目，默认20，最大值100。    |
| offset           | int          | 否  | 返回资源列表起始偏移量，默认0。         |

### 响应参数说明

| 参数名称    | 参数类型                                   | 描述   |
|---------|----------------------------------------|------|
| code    | int32                                  | 状态码  |
| message | string                                 | 请求信息 |
| data    | DescribeClusterResourcesResponseParams | 响应数据 |

### DescribeClusterResourcesResponseParams

集群资源响应信息

| 参数名称               | 参数类型                     | 描述                      |
|--------------------|--------------------------|-------------------------|
| ClusterResourceSet | array of ClusterResource | 响应数据                    |
| TotalCount         | int                      | 符合条件的总记录条数              |
| RequestId          | string                   | 唯一请求 ID，由服务端生成，每次请求都会返回 |

### ClusterResource

集群资源信息

| 参数名称           | 参数类型         | 描述                                                                 |
|----------------|--------------|--------------------------------------------------------------------|
| ClusterId      | string       | 集群唯一ID                                                             |
| Vip            | string       | IP地址                                                               |
| LoadBalancerId | string       | 负载均衡唯一ID，可能返回null，表示取不到有效值                                         |
| Idle           | string       | 资源是否闲置，可能返回null，表示取不到有效值                                           |
| ClusterName    | string       | 集群名称                                                               |
| Isp            | string       | 集群的Isp属性，如："BGP","CMCC","CUCC","CTCC","INTERNAL"，可能返回null，表示取不到有效值 |
| ClustersZone   | ClustersZone | 集群所在的可用区，可能返回null，表示取不到有效值                                         |

### ClustersZone

集群所在可用区信息

| 参数名称       | 参数类型            | 描述                          |
|------------|-----------------|-----------------------------|
| MasterZone | array of string | 集群所在的主可用区，可能返回null，表示取不到有效值 |
| SlaveZone  | array of string | 集群所在的备可用区，可能返回null，表示取不到有效值 |
