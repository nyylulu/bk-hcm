import { VendorEnum } from '@/common/constant';
export const pluginHandlerDialog = {
  vendorArr: [VendorEnum.TCLOUD, VendorEnum.ZIYAN],
};

export type PluginHandlerType = typeof pluginHandlerDialog;
