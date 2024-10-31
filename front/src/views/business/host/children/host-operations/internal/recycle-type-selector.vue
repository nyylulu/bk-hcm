<script setup lang="ts">
import { ref, watch } from 'vue';
import { ModelProperty } from '@/model/typings';

defineOptions({ name: 'recycle-type-selector' });
const props = defineProps<{ value: string }>();
const emit = defineEmits(['change']);

const selected = ref<string>(props.value);
const option = ref<ModelProperty['option']>({
  [props.value]: props.value,
  滚服类型: '滚服类型',
});

watch(selected, (v) => {
  emit('change', v);
});
</script>

<template>
  <hcm-form-enum v-model="selected" :display="{ on: 'cell' }" :option="option" />
</template>

<style scoped></style>
