import type { ICvmListOperateStatus } from '@/store/cvm-operate';

// 主机重装弹框数据
export type CvmListRestDataView = {
  reset: ICvmListOperateStatus[];
  unReset: ICvmListOperateStatus[];
  count: number;
};

export enum ImageType {
  PUBLIC_IMAGE = 'PUBLIC_IMAGE',
  PRIVATE_IMAGE = 'PRIVATE_IMAGE',
}

export interface ICvmOperateTableView {
  account_id: string;
  type: string;
  count: number;
  region: string;
  vendor: string;
  image_type: string;
  image_name: string;
  list: ICvmListOperateStatus[];
}

export interface IPreviewRecycleOrderItem {
  order_id: number;
  suborder_id: string;
  bk_biz_id: number;
  bk_biz_name: string;
  bk_username: string;
  resource_type: string;
  recycle_type: string;
  return_plan: string;
  skip_confirm: boolean;
  pool_type: number;
  cost_concerned: boolean;
  stage: string;
  status: string;
  message: string;
  handler: string;
  total_num: number;
  success_num: number;
  pending_num: number;
  failed_num: number;
  remark: string;
  create_at: string;
  update_at: string;
  sum_cpu_core: number;
  return_forecast: boolean;
  return_forecast_time: string;
}
