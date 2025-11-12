<script setup lang="ts">
import { ref, watch } from 'vue';
import http from '@/http';
import type { IDetailComponentProps } from '../../typings';
import { getMonthRange } from '../../utils';

const props = defineProps<IDetailComponentProps>();

const detailData = ref<any>({});

const fetchDetailData = async () => {
  const res = await http.post('/api/v1/woa/task/apply/analysis/average_time_consumption/compare', {
    current_date: getMonthRange(props.currentDate, 'YYYY-MM').startTime,
    compare_date: getMonthRange(props.compareDate, 'YYYY-MM').endTime,
  });
  detailData.value = res.data || {};
};

watch(
  [() => props.currentDate, () => props.compareDate],
  () => {
    fetchDetailData();
  },
  { immediate: true },
);
</script>

<template>
  <div>{{ props.currentDate }} {{ props.compareDate }} {{ detailData }}</div>
</template>
