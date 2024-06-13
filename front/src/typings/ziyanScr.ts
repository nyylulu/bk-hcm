import type { IQueryResData } from './common';
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
  organizations: string[];
  bk_biz_names: string[];
  module_names: string[];
  operators: string[];
}

export interface IDissolve {
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
