### 描述

- 该接口提供版本：v9.9.9+。
- 该接口所需权限：业务访问。
- 该接口功能描述：查询用户在当前地域支持独占集群列表。腾讯云代理接口 DescribeExclusiveClusters

### URL

POST /api/v1/cloud/vendors/tcloud/load_balancers/exclusive_clusters/describe

### 输入参数

| 参数名称             | 参数类型         | 必选 | 描述                                                                |
|------------------|--------------|----|-------------------------------------------------------------------|
| account_id       | string       | 是  | 云账户id                                                             |
| region           | string       | 是  | 地域                                                                |
| cluster_type     | string array | 否  | 按照 集群 的类型过滤，包括"TGW","STGW","VPCGW"                                |
| cluster_id       | string array | 否  | 按照 集群 的唯一ID过滤，如 ："tgw-12345678","stgw-12345678","vpcgw-12345678"。 |
| cluster_name     | string array | 否  | 按照 集群 的名称过滤。                                                      |
| cluster_tag      | string array | 否  | 按照 集群 的标签过滤。（只有TGW/STGW集群有集群标签）                                   |
| vip              | string array | 否  | 按照 集群 内的vip过滤。                                                    |
| load_balancer_id | string array | 否  | 按照 集群 内的负载均衡唯一ID过滤。                                               |
| network          | string array | 否  | 按照 集群 的网络类型过滤，如："Public","Private"。                               |
| zone             | string array | 否  | 按照 集群 所在可用区过滤，如："ap-guangzhou-1"（广州一区）。                           |
| isp              | string array | 否  | 按照TGW集群的 Isp 类型过滤，如："BGP","CMCC","CUCC","CTCC","INTERNAL"。        |
| limit            | int          | 否  | 返回可用区资源列表数目，默认20，最大值100。                                          |
| offset           | int          | 否  | 返回可用区资源列表起始偏移量，默认0。                                               |

### 响应参数说明

| 参数名称    | 参数类型                             | 描述   |
|---------|----------------------------------|------|
| code    | int32                            | 状态码  |
| message | string                           | 请求信息 |
| data    | DescribeExclusiveClusterResponse | 响应数据 |

#### DescribeResourcesResponse

| 参数名称       | 参数类型             | 描述         |
|------------|------------------|------------|
| ClusterSet | array of Cluster | 响应数据       |
| TotalCount | int              | 符合条件的总记录条数 |

#### Cluster

可用区资源

| 参数名称                     | 参数类型   | 描述                                                             |
|--------------------------|--------|----------------------------------------------------------------|
| ClusterId                | string | 集群唯一ID                                                         |
| ClusterName              | string | 集群名称                                                           |
| ClusterType	             | string | 集群类型，如TGW，STGW，VPCGW                                           |
| ClusterTag               | string | 集群标签，只有TGW/STGW集群有标签                                           |
| Zone                     | string | 集群所在可用区，如ap-guangzhou-1                                        |
| Network                  | string | 集群网络类型，如Public，Private                                         |
| MaxConn                  | int    | 最大连接数（个/秒）                                                     |
| MaxInFlow                | int    | 最大入带宽Mbps                                                      |
| MaxInPkg                 | int    | 最大入包量（个/秒）                                                     |
| MaxOutFlow               | int    | 最大出带宽Mbps                                                      |
| MaxOutPkg                | int    | 最大出包量（个/秒）                                                     |
| MaxNewConn               | int    | 最大新建连接数（个/秒）                                                   |
| ResourceCount            | int    | 集群内资源总数目                                                       |
| IdleResourceCount        | int    | 集群内空闲资源数目                                                      |
| LoadBalanceDirectorCount | int    | 集群内转发机的数目                                                      |
| Isp                      | string | 集群的Isp属性，如："BGP","CMCC","CUCC","CTCC","INTERNAL"。              |
| ClustersVersion          | string | 集群版本                                                           |
| DisasterRecoveryType     | string | 集群容灾类型，如SINGLE-ZONE，DISASTER-RECOVERY，MUTUAL-DISASTER-RECOVERY |
| Egress                   | string | 网络出口                                                           |
| IPVersion                | string | IP版本                                                           |


