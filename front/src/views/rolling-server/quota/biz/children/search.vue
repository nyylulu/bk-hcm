<script setup lang="ts">
import { ref, watch } from 'vue';
import { ModelProperty } from '@/model/typings';
import quotaBizViewProperties from '@/model/rolling-server/quota-biz.view';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import type { IBizViewSearchProps, IBizViewSearchCondition } from '../../typings';

const props = withDefaults(defineProps<IBizViewSearchProps>(), {});

const emit = defineEmits<{
  (e: 'search', condition: IBizViewSearchCondition): void;
  (e: 'reset'): void;
}>();

const fieldIds = ['quota_month', 'bk_biz_ids', 'adjust_type', 'revisers'];
const fields = fieldIds.map((id) => quotaBizViewProperties.find((prop) => prop.id === id));

const formValues = ref<IBizViewSearchCondition>({});
let conditionInitValues: IBizViewSearchCondition;

const getSearchCompProps = (field: ModelProperty) => {
  if (field.id === 'quota_month') {
    return {
      type: 'month',
      format: 'yyyy-MM',
      clearable: false,
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
  <div class="quota-biz-search">
    <grid-container layout="vertical" :column="4" :content-min-width="'1fr'" :gap="[16, 60]">
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

<style lang="scss" scoped>
.quota-biz-search {
  margin-bottom: 16px;
  position: relative;
  z-index: 3; // fix被bk-table-head遮挡

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
