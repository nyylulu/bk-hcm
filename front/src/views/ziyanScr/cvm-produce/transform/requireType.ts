import { ref } from 'vue';
import { getRequireTypes } from '@/api/host/task';
export const useRequireTypes = () => {
  const requireTypes = ref([]);
  const fetchRequireTypes = async () => {
    const res = await getRequireTypes();
    requireTypes.value = res?.data?.info || [];
  };
  fetchRequireTypes();
  const findRequireTypes = (someValue) => {
    return requireTypes.value.find((item) => {
      return Object.values(item).some((value) => value === someValue);
    });
  };
  const getTypeCn = (someValue) => {
    return findRequireTypes(someValue)?.require_name || someValue;
  };
  return { getTypeCn };
};
