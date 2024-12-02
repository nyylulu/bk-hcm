<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { IMoaVerifyResult } from './typings';
import successIcon from '@/assets/image/corret-fill.png';
import failedIcon from '@/assets/image/delete-fill.png';

defineProps<{ verifyResult: IMoaVerifyResult }>();

const { t } = useI18n();
</script>

<template>
  <!-- 结束态 -->
  <div v-if="verifyResult.status === 'finish'" class="verify-result">
    <template v-if="verifyResult.button_type === 'confirm'">
      <img :src="successIcon" alt="" />
      <span>{{ t('校验成功') }}</span>
    </template>
    <template v-else-if="verifyResult.button_type === 'cancel'">
      <img :src="failedIcon" alt="" />
      <span>{{ t('校验失败') }}</span>
    </template>
  </div>
  <!-- 报错处理 -->
  <div v-else-if="verifyResult.status === 'error'" class="verify-result">
    <img :src="failedIcon" alt="" />
    <bk-overflow-title type="tips" class="error-message">{{ verifyResult.errorMessage }}</bk-overflow-title>
  </div>
</template>

<style scoped lang="scss">
.verify-result {
  margin-left: 12px;
  display: flex;
  align-items: center;

  img {
    margin-right: 8px;
    width: 16px;
    height: 16px;
  }

  .error-message {
    max-width: calc(100% - 24px);
    color: $danger-color;
  }
}
</style>
