import { ref } from 'vue';
import { getCvmProduceOrderStatusOpts } from '@/api/host/cvm';
export const useCvmProduceStatus = () => {
  const statusList = ref([]);
  const fetchCvmProduceStatus = async () => {
    const res = await getCvmProduceOrderStatusOpts();
    statusList.value = res?.data?.info || [];
  };
  fetchCvmProduceStatus();
  const getCvmProduceStatus = (status) =>
    statusList.value.find((item) => item.status === status)?.description || status;
  return { statusList, getCvmProduceStatus };
};
