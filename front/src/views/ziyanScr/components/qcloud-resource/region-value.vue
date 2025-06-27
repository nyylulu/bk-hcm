<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useConfigQcloudResourceStore } from '@/store/config/qcloud-resource';

defineOptions({ name: 'qcloud-region-value' });

const props = defineProps<{
  value: string | string[];
}>();

const list = ref([]);

const localValue = computed(() => {
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((region) => {
    return list.value.find((item) => item.region === region)?.region_cn;
  });
  return names?.join?.(', ');
});

const configQcloudRegionStore = useConfigQcloudResourceStore();
watchEffect(async () => {
  list.value = await configQcloudRegionStore.getQcloudRegionList();
});
</script>

<template>
  {{ displayValue }}
</template>
