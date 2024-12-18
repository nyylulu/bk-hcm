<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { ReturnedWay, useRollingServerUsageStore } from '@/store';
import { convertDateRangeToObject, getDateRange } from '@/utils/search';
import { useWhereAmI } from '@/hooks/useWhereAmI';

defineOptions({ name: 'recycle-quota-tips' });

const props = defineProps<{ selections: any[]; returnedWay: ReturnedWay }>();

const { getBizsId } = useWhereAmI();
const rollingServerUsageStore = useRollingServerUsageStore();

const typeCount = computed(() => new Set(props.selections.map((item) => item.recycle_type)).size);
const totalCount = computed(() => props.selections.reduce((prev, curr) => prev + curr.total_num, 0));
const totalCpuCoreCount = computed(() => props.selections.reduce((prev, curr) => prev + curr.sum_cpu_core, 0));
const rollingServerCpuCoreCount = computed(() =>
  props.selections.reduce((prev, curr) => {
    if (curr.recycle_type === '滚服项目') return prev + curr.sum_cpu_core;
    return prev;
  }, 0),
);
const isSelectionsHasRollingServer = computed(() => props.selections.some((item) => item.recycle_type === '滚服项目'));

const shouldBeReturnedCpuCoreCount = ref<number>(0);
const getCpuCoreSummary = async () => {
  const bk_biz_ids = [getBizsId()];
  const res = await rollingServerUsageStore.getCpuCoreSummary({
    ...convertDateRangeToObject(getDateRange('last120d')),
    bk_biz_ids,
    returned_way: props.returnedWay,
  });
  shouldBeReturnedCpuCoreCount.value = res.sum_delivered_core - res.sum_returned_applied_core;
};

onMounted(() => {
  getCpuCoreSummary();
});
</script>

<template>
  <!-- eslint-disable prettier/prettier -->
  <div v-show="!rollingServerUsageStore.cpuCoreSummaryLoading" class="mt8">
    <p class="font-small flex-row">
      <span class="text-danger">注意：</span>
      已选择 {{ typeCount }} 种项目类型的资源，共计 {{ totalCount }} 台（{{ totalCpuCoreCount }}核心），请确认后点击提交。
    </p>
    <template v-if="isSelectionsHasRollingServer">
      <!-- 普通业务 -->
      <template v-if="ReturnedWay.RESOURCE_POOL !== returnedWay">
        <p class="ext-info">滚服项目：本次退回 {{ rollingServerCpuCoreCount }} 核心后，剩余待退回 {{ shouldBeReturnedCpuCoreCount - rollingServerCpuCoreCount }} 核心，本次退回前全部应退回为 {{ shouldBeReturnedCpuCoreCount }} 核心。</p>
        <p class="ext-info">其他项目：本次退回 {{ totalCpuCoreCount - rollingServerCpuCoreCount }} 核心。</p>
      </template>
      <!-- 资源池业务 -->
      <template v-else>
        <p class="ext-info">滚服项目：本次通过资源池业务退回 {{ rollingServerCpuCoreCount }} 核心后，全平台-滚服项目-剩余待退回 {{ shouldBeReturnedCpuCoreCount - rollingServerCpuCoreCount }} 核心，本次退回前全平台-滚服项目-应退回为 {{ shouldBeReturnedCpuCoreCount }} 核心。</p>
        <p class="ext-info">其他项目：本次退回 {{ totalCpuCoreCount - rollingServerCpuCoreCount }} 核心。</p>
      </template>
    </template>
  </div>
</template>

<style scoped lang="scss">
.ext-info {
  font-size: 12px;
  text-indent: 3em;
}
</style>
