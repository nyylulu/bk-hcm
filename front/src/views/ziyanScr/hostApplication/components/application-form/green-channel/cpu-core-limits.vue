<script setup lang="ts">
import { computed, onBeforeMount } from 'vue';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useGreenChannelQuotaStore } from '@/store/green-channel/quota';

defineOptions({ name: 'green-channel-cpu-core-limits' });
const props = defineProps<{
  replicasCpuCores: number;
  bizId: string | number;
}>();

const { t } = useI18n();
const greenChannelQuotaStore = useGreenChannelQuotaStore();
const { globalQuotaConfig } = storeToRefs(greenChannelQuotaStore);

const availableCpuCoreQuota = computed(() => {
  const { biz_quota = 0, sum_delivered_core = 0 } = globalQuotaConfig.value ?? {};
  return biz_quota - sum_delivered_core;
});
const cpuCoreQuota = computed(() => globalQuotaConfig.value?.biz_quota);

onBeforeMount(() => {
  greenChannelQuotaStore.getGlobalQuota(false, [Number(props.bizId)]);
});

defineExpose({ availableCpuCoreQuota });
</script>

<template>
  <ul class="wrapper">
    <li>
      <span>{{ t('本周剩余核数：') }}</span>
      <span class="number">{{ availableCpuCoreQuota }} {{ t('核') }}</span>
    </li>
    <li>
      <span>{{ t('本周限额核数：') }}</span>
      <span class="number">{{ cpuCoreQuota }} {{ t('核') }}</span>
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
