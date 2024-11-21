<script setup lang="ts">
import { ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import dayjs from 'dayjs';

import { useRollingServerQuotaStore } from '@/store/rolling-server-quota';
import { useRollingServerUsageStore } from '@/store';
import { QuotaAdjustType } from '@/views/rolling-server/typings';
import { convertDateRangeToObject, getDateRange } from '@/utils/search';

defineOptions({ name: 'rolling-server-cpu-core-limits' });
const props = defineProps<{
  bizId: string | number;
  replicasCpuCores: number;
}>();

const { t } = useI18n();
const rollingServerQuotaStore = useRollingServerQuotaStore();
const rollingServerUsageStore = useRollingServerUsageStore();

const rollingServerCpuCoreQuota = ref(0);
const availableCpuCoreQuota = ref(0);

const getCpuQuota = async (bizId: string | number) => {
  const bk_biz_ids = [Number(bizId)];
  const [bizQuotaRes, summaryRes] = await Promise.all([
    rollingServerQuotaStore.getBizQuotaList({
      bk_biz_ids,
      quota_month: dayjs().format('YYYY-MM'),
      page: { start: 0, limit: 1, count: false },
    }),
    rollingServerUsageStore.getCpuCoreSummary({
      ...convertDateRangeToObject(getDateRange('naturalMonth')), // start, end
      bk_biz_ids,
    }),
    rollingServerQuotaStore.getGlobalQuota(),
  ]);

  const { quota = 0, quota_offset = 0, adjust_type } = bizQuotaRes.list?.[0] ?? {};
  const { sum_delivered_core = 0 } = summaryRes ?? {};
  const { globalQuotaConfig } = rollingServerQuotaStore;

  // 单个业务的实际额度 = 基础额度 + 调整额度
  const bizQuota = adjust_type === QuotaAdjustType.INCREASE ? quota + quota_offset : quota - quota_offset;
  rollingServerCpuCoreQuota.value = bizQuota;

  // 可申请的核数 = min(（单个业务的实际额度 - 单个业务已交付的核数）, 剩余额度（全平台）)
  const quotaA = bizQuota - sum_delivered_core;
  const quotaB = globalQuotaConfig.global_quota - globalQuotaConfig.sum_delivered_core;
  availableCpuCoreQuota.value = Math.min(quotaA, quotaB);
};
watchEffect(() => {
  getCpuQuota(props.bizId);
});

defineExpose({ availableCpuCoreQuota });
</script>

<template>
  <ul class="wrapper">
    <li>
      <span>{{ t('剩余额度：') }}</span>
      <span class="number">{{ availableCpuCoreQuota }} {{ t('核') }}</span>
    </li>
    <li>
      <span>{{ t('额度限制：') }}</span>
      <span class="number">{{ rollingServerCpuCoreQuota }} {{ t('核') }}</span>
    </li>
    <li>
      <span>{{ t('需求核数：') }}</span>
      <span class="number">{{ replicasCpuCores }} {{ t('核') }}</span>
    </li>
  </ul>
</template>

<style scoped lang="scss">
.wrapper {
  margin-left: 24px;
  display: flex;
  align-items: center;
  gap: 12px;

  .number {
    color: $warning-color;
  }
}
</style>
