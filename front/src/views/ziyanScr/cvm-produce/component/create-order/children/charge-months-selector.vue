<script setup lang="ts">
import { computed } from 'vue';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';

interface IProps {
  requireType: number;
  disabled?: boolean;
  isGpuDeviceType?: boolean;
}

const props = withDefaults(defineProps<IProps>(), {
  disabled: false,
  isGpuDeviceType: false,
});
const model = defineModel<number>();

const { getMonthName } = useCvmChargeType();

const isRollingServer = computed(() => props.requireType === 6);
const chargeMonthsOption = computed(() => {
  let months = isRollingServer.value
    ? Array.from({ length: 48 }, (_, i) => i + 1)
    : [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48];

  // GPU机型只能选择5年套餐
  if (props.isGpuDeviceType) {
    months = [60];
  }

  return months.reduce((acc, month) => ({ ...acc, [month]: getMonthName(month) }), {});
});
</script>

<template>
  <hcm-form-enum v-model.number="model" :option="chargeMonthsOption" :filterable="false" :disabled="disabled" />
</template>

<style scoped lang="scss"></style>
