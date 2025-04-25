<script setup lang="ts">
import { computed } from 'vue';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';

interface IProps {
  requireType: number;
  disabled?: boolean;
}

const props = withDefaults(defineProps<IProps>(), {
  disabled: false,
});
const model = defineModel<number>();

const { getMonthName } = useCvmChargeType();

const isRollingServer = computed(() => props.requireType === 6);
const chargeMonthsOption = computed(() => {
  const months = isRollingServer.value
    ? Array.from({ length: 48 }, (_, i) => i + 1)
    : [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36, 48];

  return months.reduce((acc, month) => ({ ...acc, [month]: getMonthName(month) }), {});
});
</script>

<template>
  <hcm-form-enum v-model.number="model" :option="chargeMonthsOption" :filterable="false" :disabled="disabled" />
</template>

<style scoped lang="scss"></style>
