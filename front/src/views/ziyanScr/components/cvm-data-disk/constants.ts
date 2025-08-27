export enum CvmDataDiskType {
  CLOUD_SSD = 'CLOUD_SSD',
  CLOUD_PREMIUM = 'CLOUD_PREMIUM',
  LOCAL_BASIC = 'LOCAL_BASIC',
  LOCAL_SSD = 'LOCAL_SSD',
  LOCAL_NVME = 'LOCAL_NVME',
  LOCAL_NVME_BASIC = 'LOCAL_NVME_BASIC',
  LOCAL_PRO = 'LOCAL_PRO',
}

export const CVM_DATA_DISK_INFO: Record<CvmDataDiskType, { disk_name: string; min?: number; max?: number }> = {
  CLOUD_SSD: { disk_name: 'SSD云硬盘', min: 20, max: 32000 },
  CLOUD_PREMIUM: { disk_name: '高性能云盘', min: 10, max: 32000 },
  LOCAL_BASIC: { disk_name: '本地硬盘' },
  LOCAL_SSD: { disk_name: '本地SSD硬盘' },
  LOCAL_NVME: { disk_name: '本地NVME硬盘' },
  LOCAL_NVME_BASIC: { disk_name: '本地NVME硬盘' },
  LOCAL_PRO: { disk_name: '本地HDD硬盘' },
};
