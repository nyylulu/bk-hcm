export interface CreateRecallTaskModal {
  // 机型
  device_type?: string;
  // 地域
  bk_cloud_region?: string;
  // 园区
  bk_cloud_zone?: string;
  // 下架的机器数量，最大为500
  replicas?: number;
  // 要下架的固资号列表，数量最大500。当指定固资号列表进行下架时，其他字段为空
  asset_ids?: string[];
}
