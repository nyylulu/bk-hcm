import { CvmSystemDiskType } from './constants';

export interface ICvmSystemDisk {
  disk_type: CvmSystemDiskType;
  disk_size: number;
  disk_num: number;
}
