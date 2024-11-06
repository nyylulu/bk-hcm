<script setup lang="ts">
import { ref, watch, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import dayjs from 'dayjs';

import { useRollingServerQuotaStore } from '@/store/rolling-server-quota';
import { useRollingServerUsageStore } from '@/store';
import { QuotaAdjustType } from '@/views/rolling-server/typings';
import { convertDateRangeToObject, getDateRange } from '@/utils/search';

defineOptions({ name: 'CpuCorsLimits' });
const props = defineProps<{
  cloudTableData: any[];
  bizId: string | number;
}>();

const { t } = useI18n();
const rollingServerQuotaStore = useRollingServerQuotaStore();
const rollingServerUsageStore = useRollingServerUsageStore();

const rollingServerCpuQuota = ref(0);
const availableCpuQuota = ref(0);
const replicasCpuCors = ref(0);

const calcReplicasCpuCors = async (data: any[]) => {
  // 赋值
  replicasCpuCors.value = data.reduce((prev, curr) => {
    const { replicas, spec } = curr;
    const { cpu } = spec;
    return prev + replicas * cpu;
  }, 0);
};

watch(
  () => props.cloudTableData,
  (val) => {
    if (!val.length) return;
    // 计算滚服项目的需求cpu核数
    calcReplicasCpuCors(val);
  },
  { deep: true, immediate: true },
);

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
  rollingServerCpuQuota.value = bizQuota;

  // 可申请的核数 = min(（单个业务的实际额度 - 单个业务已交付的核数）, 剩余额度（全平台）)
  const quotaA = bizQuota - sum_delivered_core;
  const quotaB = globalQuotaConfig.global_quota - globalQuotaConfig.sum_delivered_core;
  availableCpuQuota.value = Math.min(quotaA, quotaB);
};
watchEffect(() => {
  getCpuQuota(props.bizId);
});

defineExpose({ availableCpuQuota, replicasCpuCors });
</script>

<template>
  <ul class="rolling-server-info">
    <li>
      <span>{{ t('剩余额度：') }}</span>
      <span class="cpu-cors">{{ availableCpuQuota }} {{ t('核') }}</span>
    </li>
    <li>
      <span>{{ t('额度限制：') }}</span>
      <span class="cpu-cors">{{ rollingServerCpuQuota }} {{ t('核') }}</span>
    </li>
    <li>
      <span>{{ t('需求核数：') }}</span>
      <span class="cpu-cors">{{ replicasCpuCors }} {{ t('核') }}</span>
    </li>
  </ul>
</template>

<style scoped lang="scss">
.rolling-server-info {
  margin-left: 24px;
  display: flex;
  align-items: center;

  li {
    margin-right: 12px;

    &:last-of-type {
      margin-right: 0;
    }

    .cpu-cors {
      color: $warning-color;
    }
  }
}
</style>
