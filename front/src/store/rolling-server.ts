import { ref } from 'vue';
import { defineStore } from 'pinia';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { enableCount } from '@/utils/search';
import { IListResData, IQueryResData, QueryBuilderType } from '@/typings';

export enum AppliedType {
  NORMAL = 'normal',
  RESOURCE_POOL = 'resource_pool',
  CVM_PRODUCT = 'cvm_product',
}

export enum ReturnedWay {
  CRP = 'crp',
  RESOURCE_POOL = 'resource_pool',
}

interface IRollingServerBaseRecordItem {
  id: string;
  bk_biz_id: number;
  order_id: string;
  suborder_id: string;
  year: string;
  month: string;
  day: string;
  creator: string;
  created_at: string;
}

export interface IRollingServerAppliedRecordItem extends IRollingServerBaseRecordItem {
  applied_type: AppliedType;
  applied_core: number;
  delivered_core: number;
}

export interface IRollingServerReturnedRecordItem extends IRollingServerBaseRecordItem {
  applied_record_id: string;
  match_applied_core: number;
  returned_way: ReturnedWay;
}

export type RollingServerRecordItem = IRollingServerAppliedRecordItem & {
  returned_records: IRollingServerReturnedRecordItem[];
  returned_core: number;
  not_returned_core: number;
  exec_rate: string;
};

export interface IRollingServerCpuCoreSummary {
  sum_delivered_core: number;
  sum_returned_applied_core: number;
}

export const useRollingServerStore = defineStore('rolling-server', () => {
  const { getBusinessApiPath } = useWhereAmI();

  const appliedRecordsListLoading = ref(false);
  const getAppliedRecordList = async (data: QueryBuilderType) => {
    appliedRecordsListLoading.value = true;
    const api = `/api/v1/woa/${getBusinessApiPath()}rolling_servers/applied_records/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [
          Promise<IListResData<IRollingServerAppliedRecordItem[]>>,
          Promise<IListResData<IRollingServerAppliedRecordItem[]>>,
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
  const getReturnedRecordList = async (data: QueryBuilderType) => {
    returnedRecordsListLoading.value = true;
    const api = `/api/v1/woa${getBusinessApiPath()}rolling_servers/returned_records/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [
          Promise<IListResData<IRollingServerReturnedRecordItem[]>>,
          Promise<IListResData<IRollingServerReturnedRecordItem[]>>,
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

  const cpuCoreSummaryLoading = ref(false);
  const getCpuCoreSummary = async (data: {
    start: { year: number; month: number; day: number };
    end: { year: number; month: number; day: number };
    bk_biz_ids?: number[];
    order_ids?: string[];
    suborder_ids?: string[];
    returned_way?: ReturnedWay;
  }) => {
    cpuCoreSummaryLoading.value = true;
    const api = `/api/v1/woa${getBusinessApiPath()}rolling_servers/cpu_core/summary`;
    try {
      const res: IQueryResData<{ details: IRollingServerCpuCoreSummary }> = await http.post(api, data);
      return res?.data?.details ?? {};
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      cpuCoreSummaryLoading.value = false;
    }
  };

  return {
    appliedRecordsListLoading,
    getAppliedRecordList,
    returnedRecordsListLoading,
    getReturnedRecordList,
    cpuCoreSummaryLoading,
    getCpuCoreSummary,
  };
});
