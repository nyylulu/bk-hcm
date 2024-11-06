<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Select, Popover } from 'bkui-vue';
import http from '@/http';
import type { IProps, OptionsType, SelectionType } from './types';

const { Option } = Select;

defineOptions({ name: 'DeviceTypeSelector' });
const props = withDefaults(defineProps<IProps>(), {
  params: () => ({}),
  multiple: false,
  disabled: false,
  isLoading: false,
  optionDisabled: () => false,
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

const handleSort = (sortFn: (a, b) => number) => {
  options.value[props.resourceType].sort(sortFn);
};

watch(
  () => props.params,
  () => {
    getOptions();
  },
  { immediate: true, deep: true },
);

defineExpose({ handleSort });
</script>

<template>
  <Select
    v-model="selected"
    clearable
    filterable
    :multiple="multiple"
    :disabled="disabled"
    :loading="loading || isLoading"
    :placeholder="placeholder"
  >
    <!-- 遍历 options 数据 -->
    <template v-for="option in options[resourceType]" :key="option.device_type">
      <!-- 判断是否需要使用 Popover 提示 -->
      <Popover
        v-if="optionDisabledTipsContent"
        :content="optionDisabledTipsContent(option)"
        :disabled="!optionDisabled(option)"
        :popover-delay="[200, 0]"
        placement="left"
      >
        <Option :id="option.device_type" :name="option.device_type" :disabled="optionDisabled(option)">
          <!-- 如果传入了具名插槽 'option'，则渲染插槽内容 -->
          <template v-if="$slots.option">
            <slot name="option" v-bind="option"></slot>
          </template>
          <!-- 否则渲染默认的 device_type -->
          <template v-else>{{ option.device_type }}</template>
        </Option>
      </Popover>

      <!-- 如果不需要 Popover 提示，直接渲染 Option -->
      <Option v-else :id="option.device_type" :name="option.device_type" :disabled="optionDisabled(option)">
        <template v-if="$slots.option">
          <slot name="option" v-bind="option"></slot>
        </template>
        <template v-else>{{ option.device_type }}</template>
      </Option>
    </template>
  </Select>
</template>

<style scoped></style>
