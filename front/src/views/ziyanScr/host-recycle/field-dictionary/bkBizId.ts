import { ref } from 'vue';
import { useAccountStore } from '@/store';

export const useBusiness = () => {
  const businesses = ref([]);
  const getBusiness = async () => {
    const accountStore = useAccountStore();
    const { data } = await accountStore.getBizList();
    businesses.value = data || [];
  };
  getBusiness();
  const getBusinessNameById = (bkBizId) => {
    return businesses.value?.find((biz) => biz.id === bkBizId)?.name || bkBizId;
  };
  return { getBusinessNameById };
};
