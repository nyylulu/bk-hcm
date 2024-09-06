import http from '@/http';
import { ref } from 'vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useAntiAffinityLevels = () => {
  const antiAffinityLevels = ref([]);

  const transform = (level) => {
    return antiAffinityLevels.value.find((item) => item.level === level)?.description || level;
  };

  const fetchAntiAffinityLevels = async () => {
    try {
      const res = await getAntiAffinityLevels();
      antiAffinityLevels.value = res.data?.info || [];
    } catch (error) {
      antiAffinityLevels.value = [];
    }
  };
  fetchAntiAffinityLevels();
  return {
    transformAntiAffinityLevels: transform,
  };
};

export const getAntiAffinityLevels = () => {
  return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/affinity`, {});
};
