import { computed, ref, watch } from 'vue';
import rollRequest from '@blueking/roll-request';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import http from '@/http';
import type { IZoneTagSelectorProps } from './index.vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

interface IZoneItem {
  id: string;
  name: string;
  name_cn: string;
}

const useList = (props: IZoneTagSelectorProps) => {
  const list = ref([]);
  const loading = ref(false);

  const filter = computed(() => {
    const result: QueryFilterType = {
      op: 'and',
      rules: [],
    };
    if (props.vendor === VendorEnum.TCLOUD) {
      result.rules = [
        {
          field: 'vendor',
          op: QueryRuleOPEnum.EQ,
          value: props.vendor,
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
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/${vendor}/regions/${region}/zones/list`,
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
    [() => props.vendor, () => props.region],
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

const dataCommon = {
  useList,
};

export type FactoryType = typeof dataCommon;

export default dataCommon;
