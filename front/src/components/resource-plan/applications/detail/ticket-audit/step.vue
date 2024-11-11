<script setup lang="ts">
import { useI18n } from 'vue-i18n';

defineProps<{
  label: string;
  title: string;
  linkUrl: string;
}>();

const { t } = useI18n();
</script>

<template>
  <section class="info-wrap">
    <h3 class="header">
      <span class="label">{{ label }}：</span>
      <span class="title">{{ title }}</span>
      <bk-link theme="primary" target="_blank" :disabled="!linkUrl" :href="linkUrl">
        <div class="flex-row align-items-center">
          {{ t('单据详情') }}
          <Share width="12" height="12" class="ml6" />
        </div>
      </bk-link>
    </h3>
    <div class="content">
      <slot name="content"></slot>
    </div>
  </section>
</template>

<style scoped lang="scss">
.info-wrap {
  padding: 0 72px;

  .header {
    margin-bottom: 16px;
    display: inline-flex;
    align-items: center;
    font-size: 14px;
    .label {
      width: 80px;
      color: $danger-color;
    }
    .title {
      margin-right: 16px;
      color: $font-deep-color;
      font-weight: 700;
    }
  }

  .content {
    margin-left: 80px;
    padding: 12px 24px 0;
    background: #f5f7fa;
  }

  &:not(:last-of-type) {
    margin-bottom: 16px;
  }
}
</style>
