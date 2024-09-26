<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Select } from 'bkui-vue';
import http from '@/http';
import type { IProps, OptionsType, SelectionType } from './types';

const { Option } = Select;

defineOptions({ name: 'DeviceTypeSelector' });
const props = withDefaults(defineProps<IProps>(), {
  params: () => ({}),
  multiple: false,
  disabled: false,
  optionDisabled: () => false,
  optionDisabledTipsContent: () => '',
  placeholder: '请选择',
  sort: () => 0,
});
const emit = defineEmits<(e: 'change', result: SelectionType) => void>();
const model = defineModel<string | string[]>();

const selected = computed({
  get() {
    return model.value;
  },
  set(val) {
    let result: SelectionType;
    const { multiple, resourceType } = props;

    if (multiple && Array.isArray(val)) {
      result = val.reduce((prev, curr) => {
        prev.push(options.value[resourceType].find((item) => item.device_type === curr));
        return prev;
      }, []);
    } else {
      result = options.value[resourceType].find((item) => item.device_type === val);
    }

    emit('change', result);
    model.value = val;
  },
});

const options = ref<OptionsType>({ cvm: [], idcpm: [] });

const loading = ref(false);
const getOptions = async () => {
  if (props.disabled) return;
  const { resourceType, params, sort } = props;
  const { require_type, region, zone, device_group, cpu, mem, disk, enable_capacity, enable_apply } = params;

  const buildRules = (fields: Array<{ field: string; value: number | string | Array<number | string> | boolean }>) => {
    return fields.reduce((prev, curr) => {
      const { field, value } = curr;
      if (Array.isArray(value) && value.length > 0) {
        prev.push({ field, operator: 'in', value });
      }
      if (!Array.isArray(value) && value) {
        prev.push({ field, operator: 'equal', value });
      }
      return prev;
    }, []);
  };

  const rules = buildRules([
    { field: 'require_type', value: require_type },
    { field: 'region', value: region },
    { field: 'zone', value: zone },
    { field: 'label.device_group', value: device_group },
    { field: 'cpu', value: cpu },
    { field: 'mem', value: mem },
    { field: 'disk', value: disk },
    { field: 'enable_capacity', value: enable_capacity },
    { field: 'enable_apply', value: enable_apply },
  ]);

  const filter = rules.length ? { condition: 'AND', rules } : undefined;

  loading.value = true;
  try {
    const url = `/api/v1/woa/config/findmany/config/${resourceType}/devicetype`;
    const data = resourceType === 'cvm' ? { filter } : {};

    const res = await http.post(url, data);
    options.value[resourceType] = res.data?.info || [];

    if (typeof sort === 'function') {
      options.value[resourceType].sort(sort);
    }
  } catch (error) {
    options.value[resourceType] = [];
  } finally {
    loading.value = false;
  }
};

watch(
  () => props.params,
  () => {
    getOptions();
  },
  { immediate: true, deep: true },
);
</script>

<template>
  <Select
    v-model="selected"
    clearable
    filterable
    :multiple="multiple"
    :disabled="props.disabled"
    :loading="loading"
    :placeholder="placeholder"
  >
    <Option
      v-for="option in options[resourceType]"
      :key="option.device_type"
      :id="option.device_type"
      :name="option.device_type"
      :disabled="props.optionDisabled(option)"
      v-bk-tooltips="{
        content: props.optionDisabledTipsContent(option),
        disabled: !props.optionDisabled(option),
        boundary: 'parent',
        delay: 200,
      }"
    />
  </Select>
</template>

<style scoped></style>
