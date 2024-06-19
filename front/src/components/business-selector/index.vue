<script lang="ts" setup>
import { computed, ref, watchEffect, defineExpose, PropType } from 'vue';
import { useAccountStore } from '@/store';
import { isEmpty } from '@/common/util';

const props = defineProps({
  modelValue: [Number, Array] as PropType<number | number[]>,
  authed: Boolean as PropType<boolean>,
  autoSelect: Boolean as PropType<boolean>,
  isAudit: Boolean as PropType<boolean>,
  multiple: Boolean as PropType<boolean>,
  clearable: Boolean as PropType<boolean>,
});
const emit = defineEmits(['update:modelValue']);

const accountStore = useAccountStore();
const businessList = ref([]);
const defaultBusiness = ref();
const loading = ref(null);

watchEffect(async () => {
  loading.value = true;
  let req = props.authed ? accountStore.getBizListWithAuth : accountStore.getBizList;
  if (props.isAudit) {
    req = accountStore.getBizAuditListWithAuth;
  }

  const res = await req();
  loading.value = false;
  businessList.value = res?.data;
  if (props.autoSelect) {
    const id = businessList.value?.[0]?.id ?? null;
    const val = props.multiple ? [id].filter((v) => v) : id;
    defaultBusiness.value = val;
    selectedValue.value = val;
  }
});

const selectedValue = computed({
  get() {
    if (!isEmpty(props.modelValue)) {
      return props.modelValue;
    }
    return props.multiple ? [] : null;
  },
  set(val) {
    emit('update:modelValue', val);
  },
});

defineExpose({
  businessList,
  defaultBusiness,
});
</script>

<template>
  <bk-select v-model="selectedValue" :multiple="multiple" filterable :loading="loading" :clearable="clearable">
    <bk-option v-for="(item, index) in businessList" :key="index" :value="item.id" :label="item.name" />
  </bk-select>
</template>
