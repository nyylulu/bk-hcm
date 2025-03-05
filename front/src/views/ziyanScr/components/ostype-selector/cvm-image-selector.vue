<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import type { IQueryResData } from '@/typings';
import http from '@/http';

interface ICvmImage {
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
}

defineOptions({ name: 'cvm-image-selector' });
const props = withDefaults(defineProps<IProps>(), {
  idKey: 'image_id',
  displayKey: 'image_name',
  multiple: false,
  clearable: true,
  filterable: true,
  disabled: false,
});
const model = defineModel<string | string[]>();

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
    options.value = res.data.info;
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
</script>

<template>
  <bk-select
    v-model="localModel"
    :loading="loading"
    :multiple="multiple"
    :clearable="clearable"
    :filterable="filterable"
    :disabled="disabled"
  >
    <bk-option v-for="(option, index) in options" :id="option[idKey]" :key="index" :name="option[displayKey]" />
  </bk-select>
</template>

<style scoped lang="scss"></style>
