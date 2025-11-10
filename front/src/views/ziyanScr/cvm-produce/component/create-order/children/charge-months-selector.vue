<script setup lang="ts">
import { computed, useAttrs } from 'vue';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';

interface IProps {
  requireType: number;
  disabled?: boolean;
  isGpuDeviceType?: boolean;
}

const model = defineModel<number>();
const props = withDefaults(defineProps<IProps>(), {
  disabled: false,
  isGpuDeviceType: false,
});
const { getMonthName } = useCvmChargeType();

const isRollingServer = computed(() => props.requireType === 6);
const chargeMonthsOption = computed(() => {
  let months = isRollingServer.value
    ? Array.from({ length: 48 }, (_, i) => i + 1)
    : [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48];

  // GPU机型只能选择6年套餐
  if (props.isGpuDeviceType) {
    months = [72];
  }

  return months.reduce((acc, month) => ({ ...acc, [month]: getMonthName(month) }), {});
});

const attrs = useAttrs();
</script>

<template>
  <hcm-form-enum
    v-model.number="model"
    :option="chargeMonthsOption"
    :filterable="false"
    :disabled="disabled"
    v-bind="attrs"
  />
</template>

<style scoped lang="scss"></style>
