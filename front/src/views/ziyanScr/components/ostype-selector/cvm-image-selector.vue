<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import type { IQueryResData } from '@/typings';
import http from '@/http';

export interface ICvmImage {
  image_id: string;
  image_name: string;
  [key: string]: string;
}

interface IProps {
  region: string[];
  idKey?: string;
  displayKey?: string;
  multiple?: boolean;
  clearable?: boolean;
  filterable?: boolean;
  disabled?: boolean;
  transform?: (options: ICvmImage[]) => ICvmImage[];
}

defineOptions({ name: 'cvm-image-selector' });
const model = defineModel<string | string[]>();
const props = withDefaults(defineProps<IProps>(), {
  idKey: 'image_id',
  displayKey: 'image_name',
  multiple: false,
  clearable: true,
  filterable: true,
  disabled: false,
});

const emit = defineEmits(['change']);

const localModel = computed({
  get() {
    if (props.multiple && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    model.value = val;
  },
});

const loading = ref(false);
const options = ref<ICvmImage[]>([]);
const getOptions = async (region: string[]) => {
  loading.value = true;
  try {
    const res: IQueryResData<{ info: ICvmImage[] }> = await http.post('/api/v1/woa/config/findmany/config/cvm/image', {
      region,
    });
    const list = res.data.info;

    if (props.transform) {
      options.value = props.transform(list);
    } else {
      options.value = list;
    }
  } catch (error) {
    console.error(error);
    options.value = [];
  } finally {
    loading.value = false;
  }
};

watchEffect(() => {
  getOptions(props.region);
});

const handleChange = (val: string | string[]) => {
  const vals = Array.isArray(val) ? val : [val];
  emit(
    'change',
    vals,
    options.value.filter((item) => vals.includes(item[props.idKey])),
  );
};
</script>

<template>
  <bk-select
    v-model="localModel"
    :loading="loading"
    :multiple="multiple"
    :clearable="clearable"
    :filterable="filterable"
    :disabled="disabled"
    @change="handleChange"
  >
    <bk-option v-for="(option, index) in options" :id="option[idKey]" :key="index" :name="option[displayKey]">
      <template v-if="$slots['option-item']">
        <slot name="option-item" :option="option"></slot>
      </template>
      <template v-else>
        <span class="bk-select-option-item">{{ option[displayKey] }}</span>
      </template>
    </bk-option>
  </bk-select>
</template>

<style scoped lang="scss"></style>
