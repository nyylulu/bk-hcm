import { useAccountStore } from '@/store';
let businesses: any[] = [];
const getBusiness = async () => {
  const accountStore = useAccountStore();
  const { data } = await accountStore.getBizListWithAuth();
  businesses = data || [];
};

getBusiness();

export const getBusinessNameById = (bkBizId) => {
  return businesses?.find((biz) => biz.id === bkBizId)?.name || bkBizId;
};
