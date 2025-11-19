import { Ref } from 'vue';
import { VendorEnum } from '@/common/constant';
import { useZoneZiyan } from './use-zone-ziyan';
import { useZoneCommon } from './use-zone-common';

export interface IZoneItem {
  id: string;
  name: string;
}

interface ZoneHookParams {
  vendor?: VendorEnum;
  resourceType: string;
  region: string;
}

export type ZoneHook = (params: ZoneHookParams) => {
  list: Ref<IZoneItem[]>;
  loading: Ref<boolean>;
};

export function useZoneFactory(vendor: VendorEnum): ZoneHook {
  switch (vendor) {
    case VendorEnum.ZIYAN: {
      return useZoneZiyan;
    }
    default: {
      return useZoneCommon;
    }
  }
}
