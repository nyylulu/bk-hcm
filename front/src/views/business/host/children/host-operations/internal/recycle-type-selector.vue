<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { ModelProperty } from '@/model/typings';
import http from '@/http';

defineOptions({ name: 'recycle-type-selector' });
const props = defineProps<{ value: string }>();
const emit = defineEmits(['change']);

const selected = ref<string>(props.value);
const option = ref<ModelProperty['option']>();

const getOption = async () => {
  const res = await http.get('/api/v1/woa/config/find/config/requirement');
  option.value = res.data?.info?.reduce((prev: any, curr: any) => {
    const { require_name } = curr;
    prev[require_name] = require_name;
    return prev;
  }, {});
};

watch(selected, (v) => {
  emit('change', v);
});

onMounted(() => {
  getOption();
});
</script>

<template>
  <hcm-form-enum v-model="selected" :display="{ on: 'cell' }" :option="option" />
</template>

<style scoped></style>
