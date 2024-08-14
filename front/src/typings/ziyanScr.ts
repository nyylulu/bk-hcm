import type { IPageQuery, IListResData, IQueryResData } from '@/typings/common';
export interface IRecycleArea {
  id: string;
  name: string;
  start_time: string;
  end_time: string;
  which_stages: number;
  recycle_type: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export interface IQueryDissolveList {
  group_ids: string[];
  bk_biz_names: string[];
  module_names: string[];
  operators: string[];
}

export interface IDissolve {
  bk_biz_id: number | string;
  bk_biz_name: string;
  module_host_count: { [key: string]: number };
  total: {
    current: {
      host_count: number | string;
      cpu_count: number;
    };
    origin: {
      host_count: number | string;
      cpu_count: number;
    };
  };
  progress: string;
}

export type IDissolveList = IQueryResData<{ items: IDissolve[] }>;

export interface IDissolveHostOriginListParam {
  organizations?: string[];
  bk_biz_names?: string[];
  module_names: string[];
  operators?: string[];
  page?: IPageQuery;
}

export interface originListResult {
  server_asset_id: string;
  ip: string;
  bk_host_outerip: string;
  app_name: string;
  module: string;
  device_type: string;
  module_name: string;
  idc_unit_name: string;
  sfw_name_version: string;
  go_up_date: string;
  raid_id: number;
  logic_area: string;
  server_operator: string;
  server_bak_operator: string;
  device_layer: string;
  cpu_score: number;
  mem_score: number;
  inner_net_traffic_score: number;
  disk_io_score: number;
  disk_util_score: number;
  is_pass: string;
  mem4linux: number;
  inner_net_traffic: number;
  outer_net_traffic: number;
  disk_io: number;
  disk_util: number;
  disk_total: number;
  max_cpu_core_amount: number;
  group_name: string;
  center: string;
}

export type IDissolveHostOriginListResult = IListResData<originListResult[]>;

export interface IDissolveHostCurrentListParam {
  organizations?: string[];
  group_names?: string[];
  bk_biz_names?: string[];
  module_names: string[];
  operators?: string[];
  page?: IPageQuery;
}

export interface CurrentListParam {
  server_asset_id: string;
  ip: string;
  bk_host_outerip: string;
  app_name: string;
  module: string;
  device_type: string;
  module_name: string;
  idc_unit_name: string;
  sfw_name_version: string;
  go_up_date: string;
  raid_id: number;
  logic_area: string;
  server_operator: string;
  server_bak_operator: string;
  device_layer: string;
  cpu_score: number;
  mem_score: number;
  inner_net_traffic_score: number;
  disk_io_score: number;
  disk_util_score: number;
  is_pass: string;
  mem4linux: number;
  inner_net_traffic: number;
  outer_net_traffic: number;
  disk_io: number;
  disk_util: number;
  disk_total: number;
  max_cpu_core_amount: number;
  group_name: string;
  center: string;
}

export type IDissolveHostCurrentListResult = IListResData<CurrentListParam[]>;

export interface IDissolveRecycledModuleListParam {
  op: 'and' | 'or';
  rules: {
    field: string;
    op: 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'in' | 'nin' | 'cs' | 'cis';
    value: boolean | number | string | (boolean | number | string)[];
  }[];
}
