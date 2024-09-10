import { VendorEnum } from '@/common/constant';
import dataCommon, { type FactoryType } from './data-common';
import dataZiyan from './data-ziyan';

export default function optionFactory(vendor?: VendorEnum): FactoryType {
  const optionMap: { [K in VendorEnum]?: FactoryType } = {
    [VendorEnum.ZIYAN]: dataZiyan,
  };
  return optionMap[vendor] ?? dataCommon;
}
