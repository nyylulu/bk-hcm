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

  const getFirstBizId = async () => {
    if (businessFullList.value.length > 0) {
      return businessFullList.value[0].id;
    }
    const list = await getFullBusiness();
    return list?.[0]?.id;
  };

  const getFirstAuthorizedBizId = async () => {
    if (businessAuthorizedList.value.length > 0) {
      return businessAuthorizedList.value[0].id;
    }
    const list = await getAuthorizedBusiness();
    return list?.[0]?.id;
  };

  return {
    businessFullList,
    businessAuthorizedList,
    businessFullListLoading,
    businessAuthorizedListLoading,
    getFullBusiness,
    getAuthorizedBusiness,
    getFirstBizId,
    getFirstAuthorizedBizId,
  };
});
