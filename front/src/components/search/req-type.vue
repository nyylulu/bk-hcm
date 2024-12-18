<script setup lang="ts">
import { useAttrs } from 'vue';
import FormReqType from '@/components/form/req-type.vue';

defineOptions({ name: 'hcm-search-req-type' });

withDefaults(defineProps<{ multiple: boolean }>(), {
  multiple: true,
});

const model = defineModel<number | number[]>();
const attrs = useAttrs();
</script>

<template>
  <form-req-type v-model="model" :multiple="multiple" v-bind="attrs" />
</template>
