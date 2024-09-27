<script setup lang="ts">
import { computed } from 'vue';
import { Button, Input } from 'bkui-vue';
import { useI18n } from 'vue-i18n';

interface IProps {
  loading?: boolean;
  disabled?: boolean;
  placeholder?: string;
}

defineOptions({ name: 'InputWithValidate' });
const props = withDefaults(defineProps<IProps>(), {
  loading: false,
  disabled: false,
  placeholder: '请输入',
});
const emit = defineEmits<(e: 'click', val: string) => void>();
const model = defineModel<string>();

const { t } = useI18n();
const value = computed({
  get() {
    return model.value || '';
  },
  set(val) {
    model.value = val;
  },
});
const computedDisabled = computed(() => {
  return props.disabled || value.value === '';
});
</script>

<template>
  <Input v-model="value" :placeholder="placeholder">
    <template #suffix>
      <Button
        theme="primary"
        class="button"
        @click="emit('click', model)"
        :loading="loading"
        :disabled="computedDisabled"
      >
        {{ t('校验') }}
      </Button>
    </template>
  </Input>
</template>

<style scoped lang="scss">
.button {
  position: relative;
  top: -1px;
  right: -1px;
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
  min-width: 88px;
}
</style>
