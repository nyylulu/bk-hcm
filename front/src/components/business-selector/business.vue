<script setup lang="ts">
import { ref, watchEffect } from 'vue';
import { useBusinessGlobalStore, type IBusinessItem } from '@/store/business-global';

export type BusinessScopeType = 'full' | 'auth';

export interface IBusinessSelectorProps {
  disabled?: boolean;
  multiple?: boolean;
  clearable?: boolean;
  filterable?: boolean;
  showSelectAll?: boolean;
  collapseTags?: boolean;
  scope?: BusinessScopeType;
  data?: IBusinessItem[];
  optionDisabled?: (item: IBusinessItem) => boolean;
}

const props = withDefaults(defineProps<IBusinessSelectorProps>(), {
  disabled: false,
  multiple: false,
  clearable: true,
  filterable: true,
  showSelectAll: false,
  scope: 'full',
});

const model = defineModel<number | number[]>();

const businessGlobalStore = useBusinessGlobalStore();

const list = ref<IBusinessItem[]>([]);
const loading = ref(false);

watchEffect(async () => {
  loading.value = true;
  if (props.data) {
    list.value = props.data.slice();
    loading.value = false;
  } else if (props.scope === 'full') {
    // 理论上可以直接使用store中的businessFullList，这里暂且还是重新获取一次
    list.value = await businessGlobalStore.getFullBusiness();
    loading.value = businessGlobalStore.businessFullListLoading;
  } else if (props.scope === 'auth') {
    list.value = await businessGlobalStore.getAuthorizedBusiness();
    loading.value = businessGlobalStore.businessAuthorizedListLoading;
  }
});
</script>

<template>
  <bk-select
    v-model="model"
    :disabled="disabled"
    :multiple="multiple"
    :filterable="filterable"
    :loading="loading"
    :clearable="clearable"
    :collapse-tags="collapseTags"
    :show-select-all="showSelectAll"
    :multiple-mode="multiple ? 'tag' : 'default'"
  >
    <bk-option
      v-for="item in list"
      :key="item.id"
      :value="item.id"
      :label="item.name"
      :disabled="optionDisabled?.(item) === true"
    />
  </bk-select>
</template>
