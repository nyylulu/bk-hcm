<script setup lang="ts">
import { useAttrs } from 'vue';
import { ModelProperty } from '@/model/typings';
import { DisplayType } from '../form/typings';
import Array from '../form/array.vue';

defineOptions({ name: 'hcm-search-array' });

withDefaults(defineProps<{ multiple: boolean; option: ModelProperty['option']; display?: DisplayType }>(), {
  multiple: false,
  option: () => ({}),
});

const model = defineModel<Array<any>>();

const attrs = useAttrs();
</script>

<template>
  <array v-model="model" :multiple="multiple" :option="option" :display="display" v-bind="attrs" />
</template>

<style scoped></style>
