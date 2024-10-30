<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { ReturnedWay, useRollingServerUsageStore } from '@/store';
import { convertDateRangeToObject, getDateRange } from '@/utils/search';
import { useWhereAmI } from '@/hooks/useWhereAmI';

defineOptions({ name: 'recycle-quota-tips' });

const props = defineProps<{ selections: any[]; returnedWay: ReturnedWay }>();

const { t } = useI18n();
const { getBizsId } = useWhereAmI();
const rollingServerUsageStore = useRollingServerUsageStore();

const recycleTypeCount = computed(() => new Set(props.selections.map((item) => item.recycle_type)).size);
const recycleTotalCount = computed(() => props.selections.reduce((prev, curr) => prev + curr.total_num, 0));
const recycleTotalCpuCoreCount = computed(() => props.selections.reduce((prev, curr) => prev + curr.sum_cpu_core, 0));

const shouldBeReturnedCpuCoreCount = ref<number>(0);
const getCpuCoreSummary = async () => {
  const bk_biz_ids = [getBizsId()];
  const res = await rollingServerUsageStore.getCpuCoreSummary({
    ...convertDateRangeToObject(getDateRange('naturalMonth')),
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
  <div v-show="!rollingServerUsageStore.cpuCoreSummaryLoading">
    <p class="font-small flex-row">
      <span class="text-danger">{{ t('注意：') }}</span>
      {{ t('已选择') }}
      <display-value :property="{ type: 'number' }" :value="recycleTypeCount" />
      {{ t('种项目类型的资源，共计') }}
      <display-value :property="{ type: 'number' }" :value="recycleTotalCount" />
      {{ t('台，请确认后点击提交。') }}
    </p>
    <p class="font-small" style="text-indent: 3em">
      {{ ReturnedWay.RESOURCE_POOL === returnedWay ? t('其中滚服项目资源池业务应退回为') : t('其中滚服项目应退回为') }}
      <display-value :property="{ type: 'number' }" :value="shouldBeReturnedCpuCoreCount" />
      {{ t('核心，本次将退回') }}
      <display-value :property="{ type: 'number' }" :value="recycleTotalCpuCoreCount" />
      {{ t('核心，剩余待退回') }}
      <display-value :property="{ type: 'number' }" :value="shouldBeReturnedCpuCoreCount - recycleTotalCpuCoreCount" />
      {{ t('核心。') }}
    </p>
  </div>
</template>

<style scoped lang="scss"></style>
