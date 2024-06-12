import { IPageQuery, IQueryResData } from '@/typings';

export interface IListTicketsParam {
  bk_biz_ids?: number[];
  expect_time_range?: {
    start: string;
    end: string;
  };
  obs_projects?: string[];
  ticket_ids?: string[];
  applicants?: string[];
  submit_time_range?: {
    start: string;
    end: string;
  };
  page: IPageQuery;
}

export type ResourcePlanIListTicketsResult = IQueryResData<IListTicketsResult>;

export interface IListTicketsResult {
  count?: number;
  detail?: {
    id: string;
    expect_time: string;
    bk_biz_id: number;
    bk_biz_name: string;
    bk_product_id: number;
    bk_product_name: string;
    plan_product_id: number;
    plan_product_name: string;
    demand_class: string;
    cpu_core: number;
    memory: number;
    disk_size: number;
    demand_week: string;
    demand_week_name: string;
    remark: string;
    applicant: string;
    submitted_at: string;
    created_at: string;
    updated_at: string;
  }[];
}

export type ResourcePlanTicketByIdResult = IQueryResData<TicketByIdResult>;

export interface TicketByIdResult {
  id: string;
  base_info: TicketBaseInfo;
  status_info: {
    status: 'todo' | 'auditing' | 'rejected' | 'done';
    status_name: string;
    itsm_sn: string;
    itsm_url: string;
    crp_sn: string;
    crp_url: string;
  };
  demands: TicketDemands[];
}

export interface TicketDemands {
  obs_project: string;
  expect_time: string;
  area_id: string;
  area_name: string;
  region_id: string;
  region_name: string;
  zone_id: string;
  zone_name: string;
  res_mode: string;
  demand_source: string;
  remark: string;
  cvm: {
    res_mode: string;
    device_family: string;
    device_type: string;
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
  applicant: string;
  bk_biz_id: number;
  bk_biz_name: string;
  bk_product_id: number;
  bk_product_name: string;
  plan_product_id: number;
  plan_product_name: string;
  virtual_dept_id: number;
  virtual_dept_name: string;
  demand_class: string;
  created_at: string;
  submitted_at: string;
  remark: string;
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
