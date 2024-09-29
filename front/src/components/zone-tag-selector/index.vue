<script lang="ts" setup>
import { computed } from 'vue';
import { useFormItem } from 'bkui-vue/lib/shared';
import { VendorEnum } from '@/common/constant';
import CollapsibleTag from '@blueking/collapsible-tag';
import '@blueking/collapsible-tag/vue3/vue3.css';
import dataFactory from './data-factory';

defineOptions({ name: 'ZoneTagSelector' });

export interface IZoneTagSelectorProps {
  vendor: VendorEnum;
  region: string;
  resourceType?: string;
  separateCampus?: boolean;
  disabled?: boolean;
  maxWidth?: number;
  minWidth?: number;
  autoExpand?: 'selected' | boolean;
  emptyText?: string;
}

const props = withDefaults(defineProps<IZoneTagSelectorProps>(), {
  disabled: false,
  maxWidth: 200,
  minWidth: 200,
  autoExpand: 'selected',
});

const emit = defineEmits<(e: 'change', value: string) => void>();

const model = defineModel<string>();
const formItem = useFormItem();

const selected = computed({
  get() {
    return [model.value];
  },
  set(val) {
    [model.value] = val;
    emit('change', val[0] ?? '');
    formItem?.validate('change');
  },
});

const { useList } = dataFactory(props.vendor);

const { list } = useList(props);
</script>

<template>
  <CollapsibleTag
    v-model="selected"
    :data="list"
    :disabled="disabled"
    :empty-text="emptyText"
    :tag-max-width="maxWidth"
    :tag-min-width="minWidth"
    :auto-expand="autoExpand"
  />
</template>
