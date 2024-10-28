import { ref } from 'vue';
import { defineStore } from 'pinia';
import dayjs from 'dayjs';
import { IListResData, IQueryResData, QueryParamsType } from '@/typings';
import rollRequest from '@blueking/roll-request';
import type { QuotaAdjustType } from '@/views/rolling-server/typings';
import { enableCount } from '@/utils/search';
import http from '@/http';

export interface IRollingServerBizQuotaItem {
  id: string;
  offset_config_id: string;
  year: number;
  month: number;
  bk_biz_id: number;
  bk_biz_name: string;
  base_quota: number;
  adjust_type: QuotaAdjustType;
  quota_offset: number;
  quota_offset_final: number;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export interface IGlobalQuota {
  id: string;
  global_quota: number;
  biz_quota: number;
  unit_price: number;
}

export interface IGlobalCpuCoreSummary {
  sum_delivered_core: number;
  sum_returned_applied_core: number;
}

export interface IAdjustRecordItem {
  id: number;
  offset_config_id: string;
  operator: string;
  adjust_type: QuotaAdjustType;
  quota_offset: number;
  created_at: string;
}

export const useRollingServerQuotaStore = defineStore('rolling-server-quota', () => {
  const bizQuotaListLoading = ref(false);
  const createBizQuotaLoading = ref(false);
  const adjustBizQuotaLoading = ref(false);
  const adjustRecordsLoading = ref(false);
  const globalQuotaConfig = ref<Partial<IGlobalQuota & IGlobalCpuCoreSummary>>({});

  const getGlobalQuota = async () => {
    const startOfMonth = dayjs().startOf('month');
    const endOfMonth = dayjs().endOf('month');
    try {
      const [globalQuotaRes, globalCpuCoreRes] = await Promise.all<
        [Promise<IQueryResData<IGlobalQuota>>, Promise<IListResData<IGlobalCpuCoreSummary>>]
      >([
        http.get('/api/v1/woa/rolling_servers/global_config'),
        http.post('/api/v1/woa/rolling_servers/cpu_core/summary', {
          start: {
            year: startOfMonth.year(),
            month: startOfMonth.month() + 1,
            day: startOfMonth.date(),
          },
          end: {
            year: endOfMonth.year(),
            month: endOfMonth.month() + 1,
            day: endOfMonth.date(),
          },
        }),
      ]);

      globalQuotaConfig.value = { ...globalQuotaRes.data, ...globalCpuCoreRes.data.details };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const getBizQuotaList = async (params: QueryParamsType & { bk_biz_id: number }) => {
    const { bk_biz_id, ...data } = params;
    bizQuotaListLoading.value = true;
    const api = `/api/v1/woa/rolling_servers/biz_quotas/list`;
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<IRollingServerBizQuotaItem[]>>, Promise<IListResData<IRollingServerBizQuotaItem[]>>]
      >([http.post(api, enableCount(data, false)), http.post(api, enableCount(data, true))]);
      const [{ details: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      bizQuotaListLoading.value = false;
    }
  };

  const createBizQuota = async (params: { bk_biz_ids: number[]; quota_month: string; quota: number }) => {
    createBizQuotaLoading.value = true;
    try {
      const res: IQueryResData<{ id: string }> = await http.post(
        '/api/v1/woa/rolling_servers/biz_quotas/batch/create',
        params,
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      createBizQuotaLoading.value = false;
    }
  };

  const adjustBizQuota = async (params: {
    bk_biz_ids: number[];
    adjust_month: {
      start: string;
      end: string;
    };
    adjust_type: QuotaAdjustType;
    quota_offset: number;
  }) => {
    adjustBizQuotaLoading.value = true;
    try {
      const res: IQueryResData<{ ids: string }> = await http.patch(
        '/api/v1/woa/rolling_servers/quota_offsets/batch',
        params,
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      adjustBizQuotaLoading.value = false;
    }
  };

  const getAdjustRecords = async (params: { offset_config_ids: string[] }) => {
    adjustRecordsLoading.value = true;
    try {
      const list = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount<IAdjustRecordItem>('/api/v1/woa/rolling_servers/quota_offsets/adjust_records/list', params, {
        limit: 500,
        countGetter: (res) => res.data.count,
        listGetter: (res) => res.data.details,
      });
      return list as IAdjustRecordItem[];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      adjustRecordsLoading.value = false;
    }
  };

  return {
    bizQuotaListLoading,
    globalQuotaConfig,
    createBizQuotaLoading,
    adjustBizQuotaLoading,
    adjustRecordsLoading,
    getGlobalQuota,
    getBizQuotaList,
    createBizQuota,
    adjustBizQuota,
    getAdjustRecords,
  };
});
