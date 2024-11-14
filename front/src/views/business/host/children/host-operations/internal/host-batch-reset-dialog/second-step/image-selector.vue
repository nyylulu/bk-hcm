<template>
  <select-column
    ref="selectRef"
    style="height: 42px"
    v-model="localValue"
    :clearable="false"
    :loading="isLoading"
    :list="selectList"
    @change="handleChange"
  />
</template>

<script setup lang="ts">
import { ref, watch, watchEffect } from 'vue';
import { SelectColumn } from '@blueking/ediatable';
import type { IListItem } from '@blueking/ediatable/typings/components/select-column.vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import rollRequest from '@blueking/roll-request';

interface Props {
  modelValue?: string;
  accountId: string;
  region: string;
  vendor: string;
  imageType: string;
}

interface Params {
  account_id: string;
  region: string;
  page?: {
    limit?: number;
    offset?: number;
  };
  cloud_ids?: string[];
  filters?: { name: string; values: string[] }[];
}

export interface IImageItem extends IListItem {
  cloud_id: string;
  name: string;
  architecture: string;
  platform: string;
  state: string;
  type: string;
  image_size: number;
  image_source: string;
  os_type: string;
}
type ImageList = IImageItem[];

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
});
const emit = defineEmits<(e: 'change', v: IImageItem) => void>();

const { getBusinessApiPath } = useWhereAmI();

const selectRef = ref();
const localValue = ref(props.modelValue);

const isLoading = ref(false);
const selectList = ref<ImageList>([]);

const handleChange = async () => {
  const value = await selectRef.value.getValue();
  const target = selectList.value.find((item) => item.cloud_id === value);
  emit('change', target);
};

const getList = async (params: Params) => {
  const list = (await rollRequest({
    httpClient: http,
    pageStartKey: 'offset',
  }).rollReq(`/api/v1/cloud/${getBusinessApiPath()}vendors/${props.vendor}/images/query_from_cloud`, params, {
    limit: 100,
    countGetter: (res) => res.data.count,
    listGetter: (res) => res.data.details,
  })) as ImageList;

  return list;
};

watch(
  () => props.imageType,
  async (imageType) => {
    if (!imageType) return;
    const { accountId, region } = props;
    const params: Params = { account_id: accountId, region, filters: [{ name: 'image-type', values: [imageType] }] };
    isLoading.value = true;
    try {
      const list = await getList(params);
      // id-key, display-key 没有透传, 暂时用默认的
      selectList.value = list.map((item) => ({ ...item, label: item.name, value: item.cloud_id }));
    } catch (error) {
      console.error(error);
      selectList.value = [];
    } finally {
      isLoading.value = false;
    }
  },
  {
    immediate: true,
  },
);

watchEffect(() => {
  localValue.value = props.modelValue;
});
</script>
