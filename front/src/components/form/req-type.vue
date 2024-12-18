<script setup lang="ts">
import { computed, ref, useAttrs, watchEffect } from 'vue';
import { useConfigRequirementStore, type IRequirementItem } from '@/store/config/requirement';

defineOptions({ name: 'hcm-form-req-type' });

const props = withDefaults(defineProps<{ multiple?: boolean; clearable?: boolean; disabled?: boolean }>(), {
  multiple: false,
});

const model = defineModel<number | number[]>();
const attrs = useAttrs();

const list = ref<IRequirementItem[]>([]);

const localModel = computed({
  get() {
    if (props.multiple && model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(value) {
    const newVal = Array.isArray(value) ? value.map((val) => Number(val)) : Number(value);
    model.value = newVal;
  },
});

const configRequirementStore = useConfigRequirementStore();

watchEffect(async () => {
  list.value = await configRequirementStore.getRequirementType();
});
</script>

<template>
  <bk-select
    v-model="localModel"
    :list="list"
    :clearable="clearable"
    :multiple="multiple"
    :multiple-mode="multiple ? 'tag' : 'default'"
    :id-key="'require_type'"
    :display-key="'require_name'"
    v-bind="attrs"
  />
</template>
