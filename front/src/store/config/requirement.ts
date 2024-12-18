import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IQueryResData } from '@/typings';

export interface IRequirementItem {
  id: number;
  position: number;
  require_name: string;
  require_type: number;
}

type RequirementResponse = IQueryResData<{ count: number; info: IRequirementItem[] }>;

export const useConfigRequirementStore = defineStore('config-requirement', () => {
  const requirementTypeList = ref<IRequirementItem[]>();

  const getRequirementType = async () => {
    if (requirementTypeList.value) {
      return requirementTypeList.value;
    }
    try {
      const res: RequirementResponse = await http.get('/api/v1/woa/config/find/config/requirement');

      const list = res?.data?.info ?? [];
      requirementTypeList.value = list;

      return requirementTypeList.value;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    getRequirementType,
  };
});
