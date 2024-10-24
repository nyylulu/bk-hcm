import { ref } from 'vue';
import { defineStore } from 'pinia';
import { IListResData, QueryBuilderType } from '@/typings';
import { enableCount } from '@/utils/search';
import http from '@/http';

export enum AppliedType {
  NORMAL = 'normal',
  RESOURCE_POOL = 'resource_pool',
  CVM_PRODUCT = 'cvm_product',
}

export enum ReturnedWay {
  CRP = 'crp',
  RESOURCE_POOL = 'resource_pool',
}

export interface IRollingServerAppliedRecordsItem {
  id: string;
  applied_type: AppliedType;
  bk_biz_id: string;
  order_id: string;
  suborder_id: string;
  year: string;
  month: string;
  day: string;
  applied_core: string;
  delivered_core: string;
  creator: string;
  created_at: string;
}

export interface IRollingServerReturnedRecordsItem {
  id: string;
  bk_biz_id: string;
  order_id: string;
  suborder_id: string;
  applied_record_id: string;
  match_applied_core: number;
  year: string;
  month: string;
  day: string;
  returned_way: ReturnedWay;
  creator: string;
  created_at: string;
}

export const useRollingServerStore = defineStore('rolling-server', () => {
  const appliedRecordsListLoading = ref(false);

  const getAppliedRecordsList = async (params: QueryBuilderType & { bk_biz_id: number }) => {
    const { bk_biz_id, ...data } = params;
    appliedRecordsListLoading.value = true;
    const api = `/api/v1/woa/rolling_servers/applied_records/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [
          Promise<IListResData<IRollingServerAppliedRecordsItem[]>>,
          Promise<IListResData<IRollingServerAppliedRecordsItem[]>>,
        ]
      >([http.post(api, enableCount(data, false)), http.post(api, enableCount(data, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      appliedRecordsListLoading.value = false;
    }
  };

  const returnedRecordsListLoading = ref(false);
  const getReturnedRecordsList = async (params: QueryBuilderType & { bk_biz_id: number }) => {
    const { bk_biz_id, ...data } = params;
    returnedRecordsListLoading.value = true;
    const api = `/api/v1/woa/rolling_servers/returned_records/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [
          Promise<IListResData<IRollingServerReturnedRecordsItem[]>>,
          Promise<IListResData<IRollingServerReturnedRecordsItem[]>>,
        ]
      >([http.post(api, enableCount(data, false)), http.post(api, enableCount(data, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      returnedRecordsListLoading.value = false;
    }
  };

  return {
    appliedRecordsListLoading,
    returnedRecordsListLoading,
    getAppliedRecordsList,
    getReturnedRecordsList,
  };
});
