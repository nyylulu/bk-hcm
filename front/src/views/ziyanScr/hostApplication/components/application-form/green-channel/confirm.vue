<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

defineOptions({ name: 'green-channel-confirm' });
const emit = defineEmits<{
  (e: 'confirm'): void;
  (e: 'cancel', oldRequireType: number): void;
}>();

const { t } = useI18n();

const isShow = ref(false);
let cacheRequireType: number;
const show = (oldRequireType: number) => {
  isShow.value = true;
  cacheRequireType = oldRequireType;
};

const isAgree = ref(false);
const handleConfirm = () => {
  emit('confirm');
  isShow.value = false;
};
const handleClosed = () => {
  emit('cancel', cacheRequireType);
  isShow.value = false;
};

defineExpose({ show });
</script>

<template>
  <bk-dialog v-model:is-show="isShow" title="请确认" width="640" @closed="handleClosed">
    <p class="tips">
      {{ t('当前选择的是') }}
      <span class="text-danger">{{ t('小额绿通') }}</span>
      <!-- eslint-disable-next-line prettier/prettier -->
      {{ t('，用于应对突发的应急资源申请需求，如故障替换，加急申请等特殊业务场景，无需预测报备可直接申请。普通业务场景的需求，请勿动用该类型申请。') }}
    </p>
    <p class="tips">
      <!-- eslint-disable-next-line prettier/prettier -->
      {{ t('该类型资源额度，为IEG BG自然月额度限制，当月的额度一旦用尽，需要紧急申请的业务将无法使用，请各业务合理使用该资源。') }}
    </p>
    <bk-checkbox v-model="isAgree">{{ t('我已知晓资源类型，并确保合理使用') }}</bk-checkbox>

    <template #footer>
      <bk-button theme="primary" @click="handleConfirm" :disabled="!isAgree">{{ t('确定') }}</bk-button>
      <bk-button @click="handleClosed" class="ml8">{{ t('取消') }}</bk-button>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.tips {
  text-indent: 2em;
  margin-bottom: 1em;
}
</style>
