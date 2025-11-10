import { type ICvmDevicetypeItem } from '@/store/cvm/device';
import { RES_ASSIGN_TYPE } from './constants';

export interface ICvmDeviceTypeFormData {
  // 机型
  deviceTypes: string[];
  // 完整的机型数据
  deviceTypeList?: ICvmDevicetypeItem[];
  zones: string[];
  chargeType: string;
  chargeMonths: number;
  resAssignType: keyof typeof RES_ASSIGN_TYPE;
  inheritAssetId?: string;
  inheritInstanceId?: string;
}
