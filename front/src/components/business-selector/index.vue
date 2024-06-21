<script lang="ts" setup>
import { computed, ref, watchEffect, defineExpose, PropType } from 'vue';
import { useAccountStore } from '@/store';
import { isEmpty } from '@/common/util';
import { useI18n } from 'vue-i18n';

const props = defineProps({
  modelValue: [Number, String, Array] as PropType<number | string | Array<number | string>>,
  authed: Boolean as PropType<boolean>,
  autoSelect: Boolean as PropType<boolean>,
  isAudit: Boolean as PropType<boolean>,
  multiple: Boolean as PropType<boolean>,
  clearable: Boolean as PropType<boolean>,
  isShowAll: Boolean as PropType<boolean>,
});
const emit = defineEmits(['update:modelValue']);

const { t } = useI18n();
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
  if (props.isShowAll) {
    businessList.value.unshift({
      name: t('全部'),
      id: 'all',
    });
  }
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
    if (props.isShowAll) {
      if (props.multiple && Array.isArray(props.modelValue) && props.modelValue.length === 0) {
        return ['all'];
      }
      if (!props.multiple && props.modelValue === '') {
        return 'all';
      }
    }
    return props.multiple ? [] : null;
  },
  set(val) {
    let selectedValue = val;
    if (props.isShowAll) {
      if (props.multiple && Array.isArray(selectedValue)) {
        if (selectedValue[selectedValue.length - 1] === 'all') {
          selectedValue = [];
        } else if (selectedValue.includes('all')) {
          const index = selectedValue.findIndex((val) => val === 'all');
          selectedValue.splice(index, 1);
        }
      }
      if (!props.multiple && selectedValue === 'all') {
        selectedValue = '';
      }
    }
    emit('update:modelValue', selectedValue);
  },
});

defineExpose({
  businessList,
  defaultBusiness,
});
</script>

<template>
  <bk-select v-model="selectedValue" :multiple="multiple" filterable :loading="loading" :clearable="clearable">
    <bk-option v-for="item in businessList" :key="item.id" :value="item.id" :label="item.name" />
  </bk-select>
</template>
