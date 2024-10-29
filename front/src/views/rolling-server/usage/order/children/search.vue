<script setup lang="ts">
import { ref, watch } from 'vue';
import type { ModelProperty } from '@/model/typings';
import type { ISearchCondition, ISearchProps } from '../../typings';
import usageOrderViewProperties from '@/model/rolling-server/usage-order.view';

import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';

const props = withDefaults(defineProps<ISearchProps>(), {});
const emit = defineEmits<{
  (e: 'search', condition: ISearchCondition): void;
  (e: 'reset'): void;
}>();

const formValues = ref<ISearchCondition>({});
let conditionInitValues: ISearchCondition;

const fieldIds = ['created_at', 'bk_biz_id', 'suborder_id'];
const fields = fieldIds.map((id) => usageOrderViewProperties.find((view) => view.id === id));

const getSearchCompProps = (field: ModelProperty) => {
  if (field.id === 'created_at') {
    return { type: 'daterange', format: 'yyyy-MM-dd' };
  }

  if (field.id === 'bk_biz_id') {
    return { showAll: true, allOptionId: -1, scope: 'auth' };
  }
  if (field.id === 'suborder_id') {
    return {
      display: { appearance: 'tag-input' },
      'paste-fn': (value: string) => value.split(',').map((tag) => ({ id: tag, name: tag })),
    };
  }

  return {
    option: field.option,
  };
};

const handleSearch = () => {
  emit('search', formValues.value);
};

const handleReset = () => {
  formValues.value = { ...conditionInitValues };
  emit('reset');
};

watch(
  () => props.condition,
  (condition) => {
    formValues.value = { ...condition };
    // 只记录第一次的condition值，重置时回到最开始的默认值
    if (!conditionInitValues) {
      conditionInitValues = { ...formValues.value };
    }
  },
  { deep: true, immediate: true },
);
</script>

<template>
  <div class="rolling-server-usage-search">
    <grid-container layout="vertical" :column="3" :content-min-width="300" :gap="[16, 60]">
      <grid-item-form-element v-for="field in fields" :key="field.id" :label="field.name">
        <component :is="`hcm-search-${field.type}`" v-bind="getSearchCompProps(field)" v-model="formValues[field.id]" />
      </grid-item-form-element>
      <grid-item :span="4" class="row-action">
        <bk-button theme="primary" @click="handleSearch">查询</bk-button>
        <bk-button @click="handleReset">重置</bk-button>
      </grid-item>
    </grid-container>
  </div>
</template>

<style scoped lang="scss">
.rolling-server-usage-search {
  padding: 16px 24px;
  margin-bottom: 16px;
  position: relative;
  z-index: 3; // fix被bk-table-head遮挡
  border-radius: 2px;
  background-color: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;

  .row-action {
    padding: 4px 0;
    :deep(.item-content) {
      gap: 10px;
    }
    .bk-button {
      min-width: 86px;
    }
  }
}
</style>
