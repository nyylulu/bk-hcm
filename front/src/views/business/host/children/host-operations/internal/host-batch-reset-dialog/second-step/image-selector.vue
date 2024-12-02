<template>
  <select-column
    ref="selectRef"
    style="height: 42px"
    id-key="cloud_id"
    display-key="name"
    v-model="localValue"
    :clearable="false"
    :loading="isLoading"
    :scroll-loading="isScrollLoading"
    :list="selectList"
    :remote-method="remoteMethod"
    @change="handleChange"
    @scroll-end="handleScrollEnd"
  />
</template>

<script setup lang="ts">
import { reactive, ref, watch, watchEffect } from 'vue';
import { debounce } from 'lodash';

import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { IListItem } from '@blueking/ediatable/typings/components/select-column.vue';
import type { IListResData } from '@/typings';
import http from '@/http';

import { SelectColumn } from '@blueking/ediatable';

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
const isScrollLoading = ref(false);
const selectList = ref<ImageList>([]);
const totalCount = ref(0);
const params = reactive<Params>({
  account_id: props.accountId,
  region: props.region,
  filters: [],
  page: { limit: 100, offset: 0 },
});

const handleChange = async () => {
  const value = await selectRef.value.getValue();
  const target = selectList.value.find((item) => item.cloud_id === value);
  emit('change', target);
};

const reset = () => {
  Object.assign(params, { filters: [], page: { limit: 100, offset: 0 } });
  selectList.value = [];
  totalCount.value = 0;
};

const getList = async () => {
  try {
    const res: IListResData<ImageList> = await http.post(
      `/api/v1/cloud/${getBusinessApiPath()}vendors/${props.vendor}/images/query_from_cloud`,
      params,
    );
    const { details: list = [], count = 0 } = res.data ?? {};
    selectList.value = [...selectList.value, ...list];
    totalCount.value = count;

    return { list, count };
  } catch (error) {
    console.error(error);
    reset();
  }
};

const handleScrollEnd = async () => {
  if (totalCount.value <= selectList.value.length || isScrollLoading.value) return;

  isScrollLoading.value = true;
  try {
    params.page.offset += params.page.limit;
    await getList();
  } finally {
    isScrollLoading.value = false;
  }
};

// 远程搜索
const remoteMethod = debounce(async (name: string) => {
  reset();
  const filters = [{ name: 'image-type', values: [props.imageType] }];
  name && filters.push({ name: 'image-name', values: [name] });
  Object.assign(params, { filters });

  isLoading.value = true;
  try {
    await getList();
  } finally {
    isLoading.value = false;
  }
}, 500);

// 镜像类型变更时，重置列表
watch(
  () => props.imageType,
  async (imageType) => {
    if (!imageType) return;
    reset();
    Object.assign(params, { filters: [{ name: 'image-type', values: [imageType] }] });

    isLoading.value = true;
    try {
      await getList();
    } finally {
      isLoading.value = false;
    }
  },
  { immediate: true },
);

watchEffect(() => {
  localValue.value = props.modelValue;
});
</script>
