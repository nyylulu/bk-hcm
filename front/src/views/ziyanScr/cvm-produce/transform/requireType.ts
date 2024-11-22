import { ref } from 'vue';
import { getRequireTypes } from '@/api/host/task';
export const useRequireTypes = () => {
  const requireTypes = ref([]);

  const fetchRequireTypes = async () => {
    const res = await getRequireTypes();
    requireTypes.value = res?.data?.info || [];
  };

  fetchRequireTypes();

  const findRequireTypes = (require_type: number) => {
    return requireTypes.value.find((item) => item.require_type === require_type);
  };

  const getTypeCn = (require_type: number) => {
    return findRequireTypes(require_type)?.require_name || require_type;
  };

  return { getTypeCn };
};
