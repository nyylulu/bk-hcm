import { computed, ref, watch } from 'vue';
import rollRequest from '@blueking/roll-request';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import http from '@/http';
import { ZoneHook } from './use-zone-factory';

interface IZoneItem {
  id: string;
  name: string;
  name_cn: string;
}

export const useZoneCommon: ZoneHook = (params) => {
  const list = ref([]);
  const loading = ref(false);

  const filter = computed(() => {
    const result: QueryFilterType = {
      op: 'and',
      rules: [],
    };
    if (params.vendor === VendorEnum.TCLOUD) {
      result.rules = [
        {
          field: 'vendor',
          op: QueryRuleOPEnum.EQ,
          value: params.vendor,
        },
        {
          field: 'state',
          op: QueryRuleOPEnum.EQ,
          value: 'AVAILABLE',
        },
      ];
    }

    return result;
  });

  const getZoneData = async (vendor: string, region: string) => {
    loading.value = true;
    try {
      const result = await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount(
        `/api/v1/cloud/vendors/${vendor}/regions/${region}/zones/list`,
        {
          filter: filter.value,
        },
        { limit: 100, countGetter: (res) => res.data.count, listGetter: (res) => res.data.details },
      );
      list.value = (result as IZoneItem[]).map((item) => ({
        id: item.id,
        name: item.name_cn || item.name,
      }));
    } catch {
      list.value = [];
    } finally {
      loading.value = false;
    }
  };

  watch(
    [() => params.vendor, () => params.region],
    ([vendor, region]) => {
      list.value = [];
      if (vendor && region) {
        getZoneData(vendor, region);
      }
    },
    { immediate: true },
  );

  return {
    list,
    loading,
  };
};
