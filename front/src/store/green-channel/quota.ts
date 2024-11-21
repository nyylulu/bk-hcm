import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IQueryResData } from '@/typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { convertDateRangeToObject, getDateRange } from '@/utils/search';

export interface IGlobalQuota {
  ieg_quota: number;
  biz_quota: number;
  audit_threshold: number;
}

export interface ICpuCoreSummary {
  sum_delivered_core: number;
}

export interface ICpuCoreSummaryParams {
  start?: {
    year: number;
    month: number;
    day: number;
  };
  end?: {
    year: number;
    month: number;
    day: number;
  };
  bk_biz_ids?: number[];
}

export const useGreenChannelQuotaStore = defineStore('green-channel-quota', () => {
  const { getBusinessApiPath } = useWhereAmI();

  const updateQuotaConfigLoading = ref(false);

  const globalQuotaConfig = ref<Partial<IGlobalQuota & ICpuCoreSummary>>({});

  const getGlobalQuota = async () => {
    try {
      const [globalQuotaRes, globalCpuCoreRes] = await Promise.all<
        [Promise<IQueryResData<IGlobalQuota>>, Promise<IQueryResData<ICpuCoreSummary>>]
      >([
        http.get('/api/v1/woa/green_channels/configs'),
        http.post(`/api/v1/woa/${getBusinessApiPath()}green_channels/cpu_core/summary`, {
          ...convertDateRangeToObject(getDateRange('naturalIsoWeek')),
        }),
      ]);

      globalQuotaConfig.value = { ...globalQuotaRes.data, ...globalCpuCoreRes.data };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const getCpuCoreSummary = async (params: ICpuCoreSummaryParams) => {
    try {
      const api = `/api/v1/woa/${getBusinessApiPath()}green_channels/cpu_core/summary`;
      const res: IQueryResData<ICpuCoreSummary> = await http.post(api, params);

      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  const updateQuotaConfig = async (data: Partial<IGlobalQuota>) => {
    updateQuotaConfigLoading.value = true;
    try {
      await http.patch('/api/v1/woa/green_channels/configs', data);

      // 成功后更新值
      for (const [key, value] of Object.entries(data)) {
        updateQuotaConfigValue(key as keyof IGlobalQuota, value);
      }
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      updateQuotaConfigLoading.value = false;
    }
  };

  const updateQuotaConfigValue = (key: keyof IGlobalQuota, value: number) => {
    globalQuotaConfig.value[key] = value;
  };

  return {
    globalQuotaConfig,
    updateQuotaConfigLoading,
    getGlobalQuota,
    getCpuCoreSummary,
    updateQuotaConfig,
    updateQuotaConfigValue,
  };
});
