<script setup lang="ts">
import { useAttrs, useTemplateRef } from 'vue';
import HcmFormOrg from '@/components/form/org.vue';

defineOptions({ name: 'hcm-search-org' });

withDefaults(defineProps<{ disabled?: boolean; multiple?: boolean }>(), {
  disabled: false,
  multiple: true,
});

const model = defineModel<number | number[] | string | string[]>();
const modelChecked = defineModel('checked');

const attrs = useAttrs();

const orgRef = useTemplateRef('orgRef');

defineExpose({
  clear: () => orgRef.value.clear(),
});
</script>

<template>
  <hcm-form-org
    ref="orgRef"
    v-model="model"
    v-model:checked="modelChecked"
    :disabled="disabled"
    :multiple="multiple"
    v-bind="attrs"
  />
</template>
