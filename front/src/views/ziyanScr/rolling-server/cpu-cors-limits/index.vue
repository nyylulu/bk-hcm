<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

defineOptions({ name: 'CpuCorsLimits' });
const props = defineProps<{
  cloudTableData: any[];
}>();

const { t } = useI18n();

const rollingServerCpuCorsLimits = ref(10000);
const replicasCpuCors = ref(0);
const isReplicasCpuCorsExceedsLimit = computed(() => replicasCpuCors.value > rollingServerCpuCorsLimits.value);

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

defineExpose({ isReplicasCpuCorsExceedsLimit });
</script>

<template>
  <ul class="rolling-server-info">
    <li>
      <span>{{ t('滚服CPU限额：') }}</span>
      <span class="cpu-cors">{{ rollingServerCpuCorsLimits }}{{ t('核') }}</span>
    </li>
    <li>
      <span>{{ t('需求核数：') }}</span>
      <span class="cpu-cors">{{ replicasCpuCors }}{{ t('核') }}</span>
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
