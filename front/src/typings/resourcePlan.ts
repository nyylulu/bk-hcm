import { IPageQuery, IQueryResData, IListResData } from '@/typings';
import { AdjustType } from './plan';

export interface IBizResourcesTicketsParam {
  ticket_ids?: string[];
  statuses?: string[];
  ticket_types?: string[];
  applicants?: string[];
  submit_time_range?: {
    start?: string;
    end?: string;
  };
  page?: IPageQuery;
}

export interface IResourcesTicketItem {
  id: string;
  bk_biz_id: number;
  bk_biz_name: string;
  op_product_id: number;
  op_product_name: string;
  plan_product_id: number;
  plan_product_name: string;
  demand_class: string;
  status: string;
  status_name: string;
  ticket_type: string;
  ticket_type_name: string;
  original_info: {
    cvm: {
      cpu_core: number;
      memory: number;
    };
    cbs: {
      disk_size: number;
    };
  };
  updated_info: {
    cvm: {
      cpu_core: number;
      memory: number;
    };
    cbs: {
      disk_size: number;
    };
  };
  applicant: string;
  remark: string;
  submitted_at: string;
  completed_at: string;
  created_at: string;
  updated_at: string;
}

export type IBizResourcesTicketsResult = IListResData<IResourcesTicketItem>;

export interface IOpResourcesTicketsParam {
  bk_biz_ids?: number[];
  op_product_ids?: number[];
  plan_product_ids?: number[];
  ticket_ids?: string[];
  statuses?: string[];
  ticket_types?: string[];
  applicants?: string[];
  submit_time_range?: {
    start?: string;
    end?: string;
  };
  page?: IPageQuery;
}

export type IOpResourcesTicketsResult = IListResData<IResourcesTicketItem>;

export type ResourcePlanTicketByIdResult = IQueryResData<TicketByIdResult>;

export interface TicketByIdResult {
  id: string;
  base_info: TicketBaseInfo;
  status_info: {
    status: string;
    status_name: string;
    itsm_sn: string;
    itsm_url: string;
    crp_sn: string;
    crp_url: string;
  };
  demands: {
    original_info: TicketDemands;
    updated_info: TicketDemands;
  }[];
}

export interface TicketDemands {
  obs_project: string;
  expect_time: string;
  region_id: string;
  zone_id: string;
  demand_res_types: string[];
  cvm: {
    res_mode: string;
    device_family: string;
    device_type: string;
    device_class: string;
    cpu_core: number;
    memory: number;
    res_pool: string;
    core_type: string;
  };
  cbs: {
    disk_type: string;
    disk_type_name: string;
    disk_io: number;
    disk_size: number;
  };
}

export interface TicketBaseInfo {
  type: string;
  type_name: string;
  applicant: string;
  bk_biz_id: number;
  bk_biz_name: string;
  op_product_id: number;
  op_product_name: string;
  plan_product_id: number;
  plan_product_name: string;
  virtual_dept_id: number;
  virtual_dept_name: string;
  remark: string;
  submitted_at: string;
}

export interface IPlanTicket {
  bk_biz_id: number;
  demand_class: string;
  demands: IPlanTicketDemand[];
  remark: string;
}

export interface IPlanTicketDemand {
  obs_project: string;
  expect_time: string;
  region_id: string;
  region_name: string;
  zone_id: string;
  zone_name: string;
  demand_source: string;
  remark?: string;
  demand_res_types: string[];
  cvm?: {
    res_mode: string;
    device_class: string;
    device_type: string;
    os: number;
    cpu_core: number;
    memory: number;
  };
  cbs?: {
    disk_type: string;
    disk_type_name: string;
    disk_io: number;
    disk_size: number;
    disk_num: number;
    disk_per_size: number;
  };
  adjustType: AdjustType;
  crp_demand_id: number;
}

export interface IBizOrgRelation {
  bk_biz_id: number;
  bk_biz_name: string;
  bk_product_id: number;
  bk_product_name: string;
  plan_product_id: number;
  plan_product_name: string;
  virtual_dept_id: number;
  virtual_dept_name: string;
}

export interface IRegion {
  region_id: string;
  region_name: string;
}

export interface IZone {
  zone_id: string;
  zone_name: string;
}

export interface IDeviceType {
  device_type: string;
  core_type: string;
  cpu_core: number;
  memory: number;
}

export interface IDiskType {
  disk_type: string;
  disk_type_name: string;
}

interface StatusListResult {
  details: {
    status: 'init' | 'auditing' | 'rejected' | 'done' | 'failed';
    status_name: string;
  }[];
}

export type IResPlanTicketStatusListResult = IQueryResData<StatusListResult>;

interface TicketTypesResult {
  details: {
    ticket_type: string;
    ticket_type_name: string;
  }[];
}

export type ITicketTypesResult = IQueryResData<TicketTypesResult>;

interface OpProductsResult {
  details: {
    op_product_id: number;
    op_product_name: string;
  }[];
}

export type IOpProductsResult = IQueryResData<OpProductsResult>;

interface PlanProductsResult {
  details: {
    plan_product_id: number;
    plan_product_name: string;
  }[];
}

export type IPlanProductsResult = IQueryResData<PlanProductsResult>;

interface BizsByOpProductResult {
  details: Array<{
    bk_biz_id: number;
    bk_biz_name: string;
  }>;
}

export type IBizsByOpProductResult = IQueryResData<BizsByOpProductResult>;
