import { VendorEnum } from '@/common/constant';
import { FilterAccountListHandler } from '../account-selector';

export const filterAccountList: FilterAccountListHandler = (accountList: any[]) => {
  return accountList.filter((item) => [VendorEnum.TCLOUD, VendorEnum.ZIYAN].includes(item.vendor));
};
