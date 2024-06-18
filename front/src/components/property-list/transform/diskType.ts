import { ref } from 'vue';
import { getDiskTypes } from '@/api/host/cvm';
export const useDiskTypes = () => {
  const diskTypeList = ref([]);
  const fetchDiskTypes = async () => {
    const res = await getDiskTypes();
    diskTypeList.value = res?.data?.info || [];
  };
  fetchDiskTypes();
  const getDiskTypesName = (type) => diskTypeList.value.find((item) => item.disk_type === type)?.disk_name || type;
  return { getDiskTypesName };
};
