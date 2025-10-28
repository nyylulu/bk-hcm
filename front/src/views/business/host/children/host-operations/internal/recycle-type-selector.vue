<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue';
import { ModelProperty } from '@/model/typings';

defineOptions({ name: 'recycle-type-selector' });
const props = defineProps<{ originValue: string }>();
const emit = defineEmits(['change']);

const selected = ref<string>(props.originValue);
const option = ref<ModelProperty['option']>({
  [props.originValue]: props.originValue,
});

watchEffect(() => {
  if (props.originValue !== '滚服项目') {
    option.value = { [props.originValue]: props.originValue, 滚服项目: '滚服项目' };
  }
});

watch(selected, (v) => {
  emit('change', v);
});

const isChanged = computed(() => selected.value !== props.originValue);
</script>

<template>
  <div :class="{ changed: isChanged }">
    <hcm-form-enum v-model="selected" :display="{ on: 'cell' }" :option="option" />
  </div>
</template>

<style lang="scss" scoped>
.changed {
  background-color: #fdf5ea;
}
</style>
