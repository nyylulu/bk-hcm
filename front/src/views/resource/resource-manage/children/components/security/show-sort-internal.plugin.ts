import { useWhereAmI } from '@/hooks/useWhereAmI';
import { VendorEnum } from '@/common/constant';

export const showSort = (vendor: string | string[]) => {
  const { isResourcePage } = useWhereAmI();
  return vendor === VendorEnum.TCLOUD || (vendor === VendorEnum.ZIYAN && !isResourcePage);
};
