export interface IPlanTicket {
  bk_biz_id: number;
  demand_class: string;
  demands: IPlanTicketDemand[];
  remark: string;
}

export interface IPlanTicketDemand {
  obs_project: string;
  expect_time: string;
  region: string;
  zone: string;
  demand_source: string;
  remark?: string;
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
