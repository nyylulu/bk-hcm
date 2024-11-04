import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData } from '@/typings';

export interface IRollingServerResPoolBusinessItem {
  id: string;
  bk_biz_id: number;
  bk_biz_name: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export const useRollingServerStore = defineStore('rolling-server', () => {
  // 资源池业务列表
  const resPollBusinessList = ref<IRollingServerResPoolBusinessItem[]>([]);
  const resPollBusinessIds = computed(() => resPollBusinessList.value.map((item) => item.bk_biz_id));
  const getResPollBusinessList = async () => {
    try {
      const res: IListResData<IRollingServerResPoolBusinessItem[]> = await http.post(
        '/api/v1/woa/metas/respool_bizs/list',
      );
      resPollBusinessList.value = res?.data?.details;
      return res?.data?.details;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    resPollBusinessList,
    resPollBusinessIds,
    getResPollBusinessList,
  };
});
