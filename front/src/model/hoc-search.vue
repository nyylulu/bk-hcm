<script setup lang="ts">
import { useAttrs } from 'vue';
import { type ModelPropertyType } from './typings';

defineProps<{ is: ModelPropertyType; display?: any }>();

const attrs = useAttrs();

const model = defineModel();
</script>
<template>
  <component :is="`hcm-search-${is}`" :display="display" v-model="model" v-bind="attrs" />
</template>
