<script setup lang="ts">
import { useAttrs } from 'vue';
import FormReqStage from '@/components/form/req-stage.vue';

defineOptions({ name: 'hcm-search-req-stage' });

withDefaults(defineProps<{ multiple: boolean }>(), {
  multiple: true,
});

const model = defineModel<string | string[]>();
const attrs = useAttrs();
</script>

<template>
  <form-req-stage v-model="model" :multiple="multiple" v-bind="attrs" />
</template>
