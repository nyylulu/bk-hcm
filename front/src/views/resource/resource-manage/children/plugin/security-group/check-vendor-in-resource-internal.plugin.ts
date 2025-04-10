import { VendorEnum } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export const checkVendorInResource = (vendor: string | string[]) => {
  const { isResourcePage } = useWhereAmI();
  return isResourcePage && VendorEnum.ZIYAN === vendor;
};
