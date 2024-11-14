import type { ICvmListRestStatus } from '@/store/cvm/reset';

export type CvmListRestStatusData = {
  reset: ICvmListRestStatus[];
  unReset: ICvmListRestStatus[];
  count: number;
};

export enum ImageType {
  PUBLIC_IMAGE = 'PUBLIC_IMAGE',
  PRIVATE_IMAGE = 'PRIVATE_IMAGE',
}

export interface ITableModel {
  account_id: string;
  type: string;
  count: number;
  region: string;
  vendor: string;
  image_type: string;
  image_name: string;
  list: ICvmListRestStatus[];
}
