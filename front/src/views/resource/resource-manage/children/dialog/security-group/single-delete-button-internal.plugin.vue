<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed, useAttrs, useTemplateRef } from 'vue';
import { Message } from 'bkui-vue';
import { useResourceStore } from '@/store';
import DeleteButton from './single-delete-button.vue';
import MoaVerifyBtn from '@/components/moa-verify/moa-verify-btn.vue';
import { MoaRequestScene } from '@/components/moa-verify/typings';

const verifyTipsRef = useTemplateRef<HTMLElement>('verify-tips');
const moaVerifyRef = useTemplateRef<InstanceType<typeof MoaVerifyBtn>>('moa-verify');

const verifyDisabled = computed(() => moaVerifyRef.value?.verifyResult?.button_type !== 'confirm');
const { t } = useI18n();
const resourceStore = useResourceStore();

const attrs = useAttrs();

const emit = defineEmits<{
  (e: 'del', sessionId: string): void;
  (e: 'success'): void;
}>();

const loading = defineModel('loading', { default: false });

const handleDelete = async () => {
  loading.value = true;
  try {
    await resourceStore.deleteBatch(
      'security_groups',
      { ids: [attrs.id], session_id: moaVerifyRef?.value?.verifyResult?.session_id },
      { globalError: false },
    );
    emit('success');
  } catch (error: any) {
    if (error.code === 2000019) {
      // MOA校验过期
      Message({ theme: 'error', message: t('MOA校验过期，请重新发起校验后操作') });
      moaVerifyRef.value?.resetVerifyResult();
    } else {
      Message({ theme: 'error', message: error.message });
    }
  } finally {
    loading.value = false;
  }
};
</script>
<template>
  <div class="delete-container">
    <moa-verify-btn
      class="verify-container"
      ref="moa-verify"
      :disabled="$attrs.disabled"
      :scene="MoaRequestScene.sg_delete"
      :res-ids="[$attrs.id as string]"
      :boundary="verifyTipsRef"
      :success-text="t('校验成功')"
    />
    <delete-button v-bind="$attrs" :disabled="verifyDisabled" :loading="loading" @del="handleDelete"></delete-button>
    <div class="verify-tips" ref="verify-tips"></div>
  </div>
</template>

<style scoped lang="scss">
.delete-container {
  display: flex;
  flex: 1;
  position: relative;
  align-items: center;
  justify-content: space-between;
  width: calc(100% - 66px);
  margin-right: auto;
}
.verify-container {
  width: calc(100% - 58px);

  :deep(.verify-result) {
    width: calc(100% - 102px);
  }
  :deep(.error-message) {
    max-width: calc(100% - 32px);
  }
}
.verify-tips {
  position: absolute;
  left: -12px;
  top: -16px;
  width: calc(100% + 90px);
  transform: translateY(-100%);
}
</style>
