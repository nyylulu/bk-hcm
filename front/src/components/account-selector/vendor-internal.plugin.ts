import { VendorEnum } from '@/common/constant';

import { vendorProperty as defaultVendorProperty } from './vendor-property';

export const vendorProperty: { [K in VendorEnum]?: { icon: any; style: any } } = {
  ...defaultVendorProperty,
  [VendorEnum.ZIYAN]: defaultVendorProperty[VendorEnum.TCLOUD],
};
