import http from '@/http';
import { onMounted, ref } from 'vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

/**
 * 获取需求类型列表
 * @returns {Promise<any>}
 */
export const getRequireTypes = (): Promise<any> => {
  return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/requirement`);
};

export const useRequireTypes = () => {
  const requireTypes = ref<any[]>([]);

  const loadRequireTypes = async () => {
    try {
      const res = await getRequireTypes();
      requireTypes.value = res.data?.info || [];
    } catch (error) {
      requireTypes.value = [];
    }
  };

  const findRequireType = (require_type: any) => {
    return requireTypes.value.find((item) => item.require_type === require_type);
  };

  const getValueCn = (require_type: any) => {
    return findRequireType(require_type)?.require_name || require_type;
  };

  onMounted(() => {
    loadRequireTypes();
  });

  return {
    transformRequireTypes: getValueCn,
  };
};
