<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useResourcePlanStore } from '@/store';

interface IProps {
  disabled?: boolean;
  clearable?: boolean;
  multiple?: boolean;
  showRollingServerProject?: boolean; // 滚服项目只有931业务可选。注意回填时要考虑是否要回填滚服项目（业务下切换业务的case）
}

const model = defineModel<string | string[]>();
const props = withDefaults(defineProps<IProps>(), {
  disabled: false,
  clearable: true,
  multiple: false,
  showRollingServerProject: false,
});
const emit = defineEmits<{
  change: [value: string | string[]];
}>();

const resourcePlanStore = useResourcePlanStore();

const list = ref<string[]>([]);
const displayList = computed(() =>
  props.showRollingServerProject ? list.value : list.value.filter((item: string) => item !== '滚服项目'),
);
const loading = ref(false);
watchEffect(async () => {
  loading.value = true;
  try {
    const res = await resourcePlanStore.getObsProjects();
    list.value = res.data?.details ?? [];
  } finally {
    loading.value = false;
  }
});

const handleChange = (value: string | string[]) => {
  emit('change', value);
};
</script>

<template>
  <bk-select v-model="model" :disabled="disabled" :multiple="multiple" :clearable="clearable" @change="handleChange">
    <bk-option v-for="item in displayList" :key="item" :id="item" :name="item" />
  </bk-select>
</template>
