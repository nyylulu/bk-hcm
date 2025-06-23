import { VendorEnum } from '@/common/constant';
import { vendorProperty as defaultVendorProperty } from './vendor-property';

const otherProperty = defaultVendorProperty.get(VendorEnum.OTHER);

defaultVendorProperty.set(VendorEnum.ZIYAN, { icon:  defaultVendorProperty.get(VendorEnum.TCLOUD)?.icon })

// 将其他供应商放置最后
defaultVendorProperty.delete(VendorEnum.OTHER);
defaultVendorProperty.set(VendorEnum.OTHER, { icon: otherProperty?.icon });

export const vendorProperty = defaultVendorProperty;
