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
