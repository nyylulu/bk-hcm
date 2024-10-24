<script setup lang="ts">
import { useAttrs } from 'vue';
import BusinessSelector from '@/components/business-selector/index.vue';
defineOptions({ name: 'hcm-search-bizs' });

const model = defineModel<string>();

const attrs = useAttrs();
</script>

<template>
  <business-selector v-model="model" v-bind="attrs"></business-selector>
</template>
