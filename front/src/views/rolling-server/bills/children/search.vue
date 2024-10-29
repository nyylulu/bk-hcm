<script setup lang="ts">
import { ref, watch } from 'vue';
import dayjs from 'dayjs';
import { ModelProperty } from '@/model/typings';
import billsViewProperties from '@/model/rolling-server/bills.view';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import type { IBillsSearchProps, IBillsSearchCondition } from '../typings';

const props = withDefaults(defineProps<IBillsSearchProps>(), {});

const emit = defineEmits<{
  (e: 'search', condition: IBillsSearchCondition): void;
  (e: 'reset'): void;
}>();

const fieldIds = ['date', 'bk_biz_id'];
const fields = fieldIds.map((id) => billsViewProperties.find((prop) => prop.id === id));

const formValues = ref<IBillsSearchCondition>({});
let conditionInitValues: IBillsSearchCondition;

const getSearchCompProps = (field: ModelProperty) => {
  if (field.id === 'date') {
    return {
      type: 'daterange',
      format: 'yyyy-MM-dd',
      disabledDate: (date: Date) => dayjs(date).isBefore(dayjs().subtract(30, 'day')) || dayjs(date).isAfter(dayjs()),
      clearable: false,
    };
  }
  if (field.id === 'bk_biz_id') {
    return {
      multiple: false,
      clearable: false,
      autoSelect: true,
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
  <div class="bills-search">
    <grid-container layout="vertical" :column="4" :content-min-width="300" :gap="[16, 60]">
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
.bills-search {
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  padding: 16px 24px;
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
