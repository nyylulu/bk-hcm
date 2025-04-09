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

export interface IRequirementObsProject {
  [key: number]: string;
}

export enum RequirementType {
  Regular = 1,
  Spring = 2,
  Dissolve = 3,
  RollServer = 6,
  GreenChannel = 7,
  SpringResPool = 8,
}

type RequirementResponse = IQueryResData<{ count: number; info: IRequirementItem[] }>;
type RequirementObsProjectResponse = IQueryResData<IRequirementObsProject>;

export const useConfigRequirementStore = defineStore('config-requirement', () => {
  const requirementTypeList = ref<IRequirementItem[]>();
  const requirementObsProjectMap = ref<IRequirementObsProject>();

  const getRequirementType = async () => {
    if (requirementTypeList.value) {
      return requirementTypeList.value;
    }
    try {
      const res: RequirementResponse = await http.get('/api/v1/woa/config/find/config/requirement');

      const list = res?.data?.info ?? [];
      requirementTypeList.value = list.sort((a, b) => a.position - b.position);

      return requirementTypeList.value;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const getRequirementObsProject = async () => {
    if (requirementObsProjectMap.value) {
      return requirementObsProjectMap.value;
    }
    try {
      const res: RequirementObsProjectResponse = await http.post('/api/v1/woa/meta/requirement/obs_project/list');
      requirementObsProjectMap.value = res?.data ?? [];

      return requirementObsProjectMap.value;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    getRequirementType,
    getRequirementObsProject,
  };
});
