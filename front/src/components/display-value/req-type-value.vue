<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useConfigRequirementStore, type IRequirementItem } from '@/store/config/requirement';
import { AppearanceType, DisplayType } from './typings';
import Tag from './appearance/tag.vue';

const props = defineProps<{ value: number | number[]; display?: DisplayType }>();

const displayOn = computed(() => props.display?.on || 'cell');
const appearance = computed(() => props.display?.appearance);

const list = ref<IRequirementItem[]>([]);

const localValue = computed(() => {
  return Array.isArray(props.value) ? props.value : [props.value];
});

const displayValue = computed(() => {
  const names = localValue.value.map((id) => {
    return list.value.find((item) => item.id === id)?.require_name;
  });
  return names?.join?.(', ');
});

const appearanceComps: Partial<Record<AppearanceType, typeof Tag>> = {
  tag: Tag,
};

const configRequirementStore = useConfigRequirementStore();

watchEffect(async () => {
  list.value = await configRequirementStore.getRequirementType();
});
</script>

<template>
  <component
    :is="appearanceComps[appearance]"
    v-if="appearance"
    :display-value="displayValue"
    :display-on="displayOn"
    :value="value"
    v-bind="$attrs"
  />
  <span v-else>{{ displayValue }}</span>
</template>
