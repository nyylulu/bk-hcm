import http from '@/http';
import { onMounted, ref } from 'vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useApplyStages = () => {
  const stages = ref([]);

  const transform = (stage: string) => {
    return (
      stages.value.find((item: { stage: string; description: string }) => item.stage === stage)?.description || stage
    );
  };

  const fetchStages = async () => {
    try {
      const res = await getApplyStages();
      stages.value = res.data?.info || [];
    } catch (error) {
      stages.value = [];
    }
  };

  onMounted(() => {
    fetchStages();
  });

  return {
    transformApplyStages: transform,
  };
};

export const getApplyStages = () => {
  return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/apply/stage`);
};
