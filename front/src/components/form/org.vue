<script setup lang="ts">
import { useAttrs, useTemplateRef } from 'vue';
import OrgSelector from '@blueking/org-selector';
import '@blueking/org-selector/vue3/vue3.css';

defineOptions({ name: 'hcm-form-org' });

withDefaults(defineProps<{ disabled?: boolean; multiple?: boolean }>(), {
  disabled: false,
  multiple: false,
});

const model = defineModel<number | number[] | string | string[]>();
const modelChecked = defineModel('checked');

const dataUrl = `${window.PROJECT_CONFIG.BK_COMPONENT_API_URL}/api/c/compapi/v2/usermanage/fe_list_departments/`;

const attrs = useAttrs();

const orgRef = useTemplateRef('orgRef');

defineExpose({
  clear: () => orgRef.value.clear(),
});
</script>

<template>
  <org-selector
    ref="orgRef"
    v-model="model"
    v-model:checked="modelChecked"
    :disabled="disabled"
    :multiple="multiple"
    :data-url="dataUrl"
    v-bind="attrs"
  />
</template>
