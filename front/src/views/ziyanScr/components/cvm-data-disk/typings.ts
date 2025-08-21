import { CvmDataDiskType } from './constants';

export interface ICvmDataDisk {
  disk_type: CvmDataDiskType;
  disk_size: number;
  disk_num: number;
}

export interface ICvmDataDiskOption {
  disk_type: CvmDataDiskType;
  disk_name: string;
}
