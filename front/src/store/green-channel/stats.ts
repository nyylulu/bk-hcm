import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryParamsType } from '@/typings';
import { enableCount } from '@/utils/search';

export interface IStatsItem {
  bk_biz_id: number;
  order_count: number;
  sum_delivered_core: number;
  sum_applied_core: number;
}

export const useGreenChannelStatsStore = defineStore('green-channel-stats', () => {
  const statsListLoading = ref(false);

  const getStatsList = async (params: QueryParamsType) => {
    statsListLoading.value = true;
    const api = '/api/v1/woa/green_channels/statistical_record/list';
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IStatsItem[]>>, Promise<IListResData<IStatsItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list: list || [], count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      statsListLoading.value = false;
    }
  };

  return {
    statsListLoading,
    getStatsList,
  };
});
