<script setup lang="ts">
import { computed } from 'vue';
import { DisplayType } from './typings';

const props = defineProps<{ value: string | number; display: DisplayType }>();

const displayValue = computed(() => {
  if (props.display.format) {
    return props.display.format(props.value);
  }
  const values = Array.isArray(props.value) ? props.value : [props.value];
  return values?.join?.(', ') || '--';
});
</script>

<template>
  {{ displayValue }}
</template>
