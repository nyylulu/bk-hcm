import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IQueryResData } from '@/typings';
import { resolveBizApiPath } from '@/utils/search';

export interface ICpuCoreSummary {
  total_core: number;
  delivered_core: number;
}

export const useDissolveQuotaStore = defineStore('dissolve-quota', () => {
  const cpuCoreSummaryLoading = ref(false);
  const getCpuCoreSummary = async (bizId: number, params: { bk_biz_id?: number } = {}) => {
    cpuCoreSummaryLoading.value = true;
    try {
      const api = `/api/v1/woa/${resolveBizApiPath(bizId)}dissolve/cpu_core/summary`;
      const res: IQueryResData<ICpuCoreSummary> = await http.post(api, params);
      return res?.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      cpuCoreSummaryLoading.value = false;
    }
  };

  return {
    cpuCoreSummaryLoading,
    getCpuCoreSummary,
  };
});
