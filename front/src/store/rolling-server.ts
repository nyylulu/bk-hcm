import { ref } from 'vue';
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
  const resPollBizsList = ref<IRollingServerResPoolBusinessItem[]>([]);
  const getResPollBusinessList = async () => {
    try {
      const res: IListResData<IRollingServerResPoolBusinessItem[]> = await http.post(
        '/api/v1/woa/rolling_servers/respool_bizs/list',
      );
      resPollBizsList.value = res?.data?.details;
      return res?.data?.details;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    }
  };

  return {
    resPollBizsList,
    getResPollBusinessList,
  };
});
