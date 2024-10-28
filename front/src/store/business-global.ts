import { ref } from 'vue';
import { defineStore } from 'pinia';
import { IQueryResData } from '@/typings';
import http from '@/http';

export interface IBusinessItem {
  id: number;
  name: string;
}

export const useBusinessGlobalStore = defineStore('businessGlobal', () => {
  const businessFullList = ref<IBusinessItem[]>([]);
  const businessAuthorizedList = ref<IBusinessItem[]>([]);
  const businessFullListLoading = ref(false);
  const businessAuthorizedListLoading = ref(false);

  const getFullBusiness = async () => {
    businessFullListLoading.value = true;
    try {
      const { data: list = [] }: IQueryResData<IBusinessItem[]> = await http.post('/api/v1/web/bk_bizs/list');
      businessFullList.value = list;
      return list;
    } finally {
      businessFullListLoading.value = false;
    }
  };

  const getAuthorizedBusiness = async () => {
    businessAuthorizedListLoading.value = true;
    try {
      const { data: list = [] }: IQueryResData<IBusinessItem[]> = await http.post('/api/v1/web/authorized/bizs/list');
      businessAuthorizedList.value = list;
      return list;
    } finally {
      businessAuthorizedListLoading.value = false;
    }
  };

  return {
    businessFullList,
    businessAuthorizedList,
    businessFullListLoading,
    businessAuthorizedListLoading,
    getFullBusiness,
    getAuthorizedBusiness,
  };
});
