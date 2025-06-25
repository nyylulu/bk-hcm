import { ref } from 'vue';
import { defineStore } from 'pinia';
import http from '@/http';
import { IQueryResData } from '@/typings';

export interface IQcloudRegionItem {
  id: number;
  region: string;
  region_cn: string;
  cmdb_region_name: string;
}

export interface IQcloudZoneItem {
  id: number;
  zone: string;
  zone_cn: string;
  region: string;
  region_cn: string;
  cmdb_region_name: string;
  cmdb_zone_id: number;
  cmdb_zone_name: string;
}

type QcloudRegionResponse = IQueryResData<{ count: number; info: IQcloudRegionItem[] }>;
type QcloudZoneResponse = IQueryResData<{ count: number; info: IQcloudZoneItem[] }>;

export const useConfigQcloudResourceStore = defineStore('config-qcloud-resource', () => {
  const requestQueue = new Map<string, any>();

  const qcloudRegionList = ref<IQcloudRegionItem[]>();
  const qcloudRegionListLoading = ref(false);
  const getQcloudRegionList = async () => {
    const reqKey = 'config-qcloud-region';

    if (qcloudRegionList.value) {
      return qcloudRegionList.value;
    }

    if (requestQueue.has(reqKey)) {
      return requestQueue.get(reqKey) as Promise<IQcloudRegionItem[]>;
    }

    qcloudRegionListLoading.value = true;
    const requestPromise = new Promise<IQcloudRegionItem[]>(async (resolve, reject) => {
      try {
        const res: QcloudRegionResponse = await http.get('/api/v1/woa/config/find/config/qcloud/region');

        const list = res?.data?.info ?? [];
        qcloudRegionList.value = list;

        resolve(list);
      } catch (error) {
        reject(error);
      } finally {
        requestQueue.delete(reqKey);
        qcloudRegionListLoading.value = false;
      }
    });

    requestQueue.set(reqKey, requestPromise);

    return requestPromise;
  };

  const qcloudZoneListCache = new Map<string, IQcloudZoneItem[]>();
  const qcloudZoneListLoading = ref(false);
  const getQcloudZoneList = async (region: string[] = []) => {
    const reqKey = 'config-qcloud-zone'.concat(JSON.stringify(region));

    if (qcloudZoneListCache.has(reqKey)) {
      return qcloudZoneListCache.get(reqKey);
    }

    if (requestQueue.has(reqKey)) {
      return requestQueue.get(reqKey) as Promise<IQcloudZoneItem[]>;
    }

    qcloudZoneListLoading.value = true;
    const requestPromise = new Promise<IQcloudZoneItem[]>(async (resolve, reject) => {
      try {
        const res: QcloudZoneResponse = await http.post('/api/v1/woa/config/findmany/config/qcloud/zone', { region });

        const list = res?.data?.info ?? [];

        qcloudZoneListCache.set(reqKey, list);
        resolve(list);
      } catch (error) {
        reject(error);
      } finally {
        requestQueue.delete(reqKey);
        qcloudZoneListLoading.value = false;
      }
    });

    requestQueue.set(reqKey, requestPromise);

    return requestPromise;
  };

  return {
    qcloudRegionList,
    qcloudRegionListLoading,
    getQcloudRegionList,
    qcloudZoneListLoading,
    getQcloudZoneList,
  };
});
