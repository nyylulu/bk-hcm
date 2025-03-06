import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IListResData, QueryBuilderType } from '@/typings';
import { enableCount } from '@/utils/search';
import { RequirementType } from '@/store/config/requirement';

export interface ICvmDeviceItem {
  id: number;
  require_type: RequirementType;
  region: string;
  zone: string;
  device_type: string;
  cpu: number;
  mem: number;
  disk: number;
  remark: string;
  label: {
    device_group: string;
    device_size: string;
  };
  capacity_flag: number;
  enable_capacity: boolean;
  enable_apply: boolean;
  score: number;
  comment: string;
  [k: string]: any;
}

export const useCvmDeviceStore = defineStore('cvm-device', () => {
  const deviceListLoading = ref(false);

  const getDeviceList = async (params: QueryBuilderType) => {
    deviceListLoading.value = true;
    const api = '/api/v1/woa/config/findmany/config/cvm/device/detail';
    try {
      const [listRes, countRes] = await Promise.all<
        [Promise<IListResData<ICvmDeviceItem[]>>, Promise<IListResData<ICvmDeviceItem[]>>]
      >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
      const [{ info: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
      return { list, count };
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      deviceListLoading.value = false;
    }
  };

  return {
    deviceListLoading,
    getDeviceList,
  };
});
