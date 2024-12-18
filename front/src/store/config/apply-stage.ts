import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IQueryResData } from '@/typings';

export interface IApplyStageItem {
  description: string;
  stage: string;
}

type ApplyStageResponse = IQueryResData<{ info: IApplyStageItem[] }>;

export const useConfigApplyStageStore = defineStore('config-apply-stage', () => {
  const applyStageList = ref<IApplyStageItem[]>();

  const getApplyStage = async () => {
    if (applyStageList.value) {
      return applyStageList.value;
    }
    try {
      const res: ApplyStageResponse = await http.get('/api/v1/woa/config/find/config/apply/stage');

      const list = res?.data?.info ?? [];
      applyStageList.value = list;

      return applyStageList.value;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    getApplyStage,
  };
});
