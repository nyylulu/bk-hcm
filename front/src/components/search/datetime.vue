<script setup lang="ts">
import { useAttrs } from 'vue';
import DateTime from '@/components/form/datetime.vue';
import type { DatePickerValueType } from 'bkui-vue/lib/date-picker/interface';

defineOptions({ name: 'hcm-search-datetime' });

withDefaults(
  defineProps<{ format: string; type: 'date' | 'daterange' | 'datetime' | 'datetimerange' | 'month' | 'year' }>(),
  {
    format: 'yyyy-MM-dd HH:mm:ss',
    type: 'datetimerange',
  },
);

const model = defineModel<DatePickerValueType>();

const attrs = useAttrs();
</script>

<template>
  <DateTime v-model="model" v-bind="attrs" :type="type" :format="format" />
</template>
