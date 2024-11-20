<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import wName from '@/components/w-name';

defineOptions({ name: 'green-channel-tips-alert' });

const { t } = useI18n();
</script>

<template>
  <bk-alert>
    <p>
      {{
        t(
          '注意：当前选择的是小额绿通，用于突发少量资源申领、紧急申领的场景，无需提前报需求预测。每单申领 100 核心以下时不需要审批，大于 100 核心时需要业务 leader 及资源团队审批。',
        )
      }}
    </p>
    <p>
      {{ t('如有疑问，请咨询') }}
      <w-name name="ICR" :alias="t('ICR(IEG资源服务助手)')" />
      {{ t('确认。') }}
    </p>
  </bk-alert>
</template>
