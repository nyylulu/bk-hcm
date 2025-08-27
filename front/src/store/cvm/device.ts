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
  device_type_class: 'SpecialType' | 'CommonType';
  [k: string]: any;
}

export interface ICvmDevicetypeItem {
  device_type: string;
  device_type_class: string;
  device_group: string;
  cpu_amount: number;
  ram_amount: number;
  core_type: number;
  device_class: string;
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

  const devicetypeListLoading = ref(false);
  const getDevicetypeListWithoutPage = async (params: QueryBuilderType) => {
    devicetypeListLoading.value = true;
    const api = '/api/v1/woa/config/findmany/config/cvm/devicetype';
    try {
      const { page } = params;
      if (page) {
        const [listRes, countRes] = await Promise.all<
          [Promise<IListResData<ICvmDevicetypeItem[]>>, Promise<IListResData<ICvmDevicetypeItem[]>>]
        >([http.post(api, enableCount(params, false)), http.post(api, enableCount(params, true))]);
        const [{ info: list = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];
        return { list, count };
      }

      const res = await http.post(api, params);
      return res?.data?.info ?? [];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      devicetypeListLoading.value = false;
    }
  };

  return {
    deviceListLoading,
    getDeviceList,
    devicetypeListLoading,
    getDevicetypeListWithoutPage,
  };
});
