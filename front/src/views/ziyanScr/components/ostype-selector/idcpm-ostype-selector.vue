<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import type { IQueryResData } from '@/typings';
import http from '@/http';

defineOptions({ name: 'idcpm-ostype-selector' });
const props = withDefaults(defineProps<{ multiple?: boolean; clearable?: boolean; filterable?: boolean }>(), {
  multiple: false,
  clearable: true,
  filterable: true,
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
const options = ref<string[]>([]);
const getOptions = async () => {
  loading.value = true;
  try {
    const res: IQueryResData<{ info: string[] }> = await http.get('/api/v1/woa/config/find/config/idcpm/ostype');
    options.value = res.data.info;
  } catch (error) {
    console.error(error);
    options.value = [];
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  getOptions();
});
</script>

<template>
  <bk-select
    v-model="localModel"
    :loading="loading"
    :multiple="multiple"
    :clearable="clearable"
    :filterable="filterable"
  >
    <bk-option v-for="(option, index) in options" :id="option" :key="index" :name="option" />
  </bk-select>
</template>

<style scoped lang="scss"></style>
