import type { IPageQuery } from '@/typings/common';

export interface cvmProduceQueryReq {
  order_id?: number[];
  task_id?: string[];
  bk_username?: string[];
  require_type?: number[];
  device_type?: string[];
  region?: string[];
  zone?: string[];
  status?: number[];
  start?: string[];
  end?: string[];
  page: IPageQuery;
}

interface ruleObj {
  field: string;
  operator: string;
  value: any;
}
interface filterObj {
  condition: 'AND' | 'OR';
  rules: ruleObj[];
}

export interface cvmDeviceTypeReq {
  filter?: filterObj;
}

export interface cvmDeviceListReq {
  filter?: filterObj;
  page: IPageQuery;
}

export interface maxResourceCapacity {
  device_type: string;
  require_type: number;
  region: string;
  zone: string;
  vpc?: string;
  subnet?: string;
  charge_type?: string;
}

export interface deviceConfigDetail {
  filter?: filterObj;
}

interface specObj {
  region: string;
  zone: string;
  device_type: string;
  image_id: string;
  disk_size: number;
  disk_type: string;
  network_type: string;
  device_group?: any;
  vpc?: string;
  subnet?: string;
}

export interface createCvmOrder {
  bk_biz_id: number;
  bk_module_id: number;
  bk_username: string;
  require_type: number;
  remark: string;
  replicas: number;
  spec: specObj;
}
