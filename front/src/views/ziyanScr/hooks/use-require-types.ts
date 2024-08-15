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

  const findRequireType = (someValue: any) => {
    return requireTypes.value.find((item) => {
      return Object.values(item).some((value: any) => value.require_type === someValue);
    });
  };

  const getValueCn = (someValue: any) => {
    return findRequireType(someValue)?.require_name || someValue;
  };

  onMounted(() => {
    loadRequireTypes();
  });

  return {
    transformRequireTypes :getValueCn,
  };
};