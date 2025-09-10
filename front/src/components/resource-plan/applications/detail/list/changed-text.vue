<script setup lang="ts">
import { Info } from 'bkui-vue/lib/icon';
// import { Popover } from 'bkui-vue';
import { getValueByKey } from '@/common/util';
import { computed } from 'vue';

interface Props {
  colData: { updated_info?: object; original_info?: object };
  field?: string;
  ticketType: string; // 单据类型
}
const props = withDefaults(defineProps<Props>(), {
  field: '',
  colData: () => ({}),
});
const specialType = ['transfer', 'delete', 'cancel'];

const purefieldKey = computed(() => {
  return props.field.replaceAll('updated_info.', '').replaceAll('original_info.', '');
});

const originalVal = computed(() => {
  return getValueByKey(props.colData?.original_info, purefieldKey.value);
});

const updatedVal = computed(() => {
  return getValueByKey(props.colData?.updated_info, purefieldKey.value);
});

const isSpecialType = computed(() => {
  return specialType.includes(props.ticketType);
});

const isChanged = computed(() => {
  if (isSpecialType.value) {
    return false;
  }
  return originalVal.value !== updatedVal.value && !!props.colData?.original_info;
});

const content = computed(() => {
  return isChanged.value ? `修改前: ${originalVal.value}` : `暂无修改前数据`;
});

const text = computed(() => {
  if (isSpecialType.value) {
    return originalVal.value || updatedVal.value;
  }
  return updatedVal.value;
});
</script>

<template>
  <div
    class="resource-plan-detail-cell"
    v-bk-tooltips="{
      content: content,
      disabled: isSpecialType,
    }"
  >
    <Info v-if="isChanged" class="resource-plan-detail-info resource-plan-detail-text" />
    <span :class="{ 'resource-plan-detail-text': isChanged }">{{ text || '--' }}</span>
  </div>
</template>

<style lang="scss" scoped>
.resource-plan-detail-cell {
  display: flex;
  cursor: pointer;
  align-items: center;
}

.resource-plan-detail-text {
  color: #e9a24c;
}

.resource-plan-detail-info {
  font-size: 18px;
  margin-right: 4px;
}
</style>
