<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <!-- eslint-disable prettier/prettier -->
  <bk-alert class="host-apply-tips-alert" theme="warning">
    <p>{{ t('短租项目使用流程') }}</p>
    <ul>
      <ol>
        <span class="label">1.{{ t('报备资源预测：') }}</span>
        <span class="content">{{ t('业务提交预测需求，可以在13周内，并选择短租时长为1-3个月') }}</span>
      </ol>
      <ol>
        <span class="label">2.{{ t('申领主机：') }}</span>
        <span class="content">{{ t('主机申领关联资源预测需求，无预测时将不允许提单申领主机') }}</span>
      </ol>
      <ol>
        <span class="label">3.{{ t('交付主机：') }}</span>
        <span class="content">{{ t('主机默认交付到配置平台的空闲机模块，业务可按需分配到业务模块') }}</span>
      </ol>
    </ul>
  </bk-alert>
</template>

<style scoped lang="scss">
.host-apply-tips-alert {
  .label {
    font-weight: 700;
  }
}
</style>
