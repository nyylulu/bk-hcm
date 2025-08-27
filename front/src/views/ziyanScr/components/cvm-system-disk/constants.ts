export enum CvmSystemDiskType {
  CLOUD_SSD = 'CLOUD_SSD',
  CLOUD_PREMIUM = 'CLOUD_PREMIUM',
  LOCAL_BASIC = 'LOCAL_BASIC',
  LOCAL_SSD = 'LOCAL_SSD',
}

export const CVM_SYSTEM_DISK_INFO: Record<CvmSystemDiskType, { disk_name: string; min?: number; max?: number }> = {
  CLOUD_SSD: { disk_name: 'SSD云硬盘', min: 50, max: 1000 },
  CLOUD_PREMIUM: { disk_name: '高性能云盘', min: 50, max: 1000 },
  LOCAL_BASIC: { disk_name: '本地盘' },
  LOCAL_SSD: { disk_name: '本地SSD盘' },
};
