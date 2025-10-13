<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import wName from '@/components/w-name';

defineOptions({ name: 'green-channel-tips-alert' });

const { t } = useI18n();
</script>

<template>
  <bk-alert theme="warning">
    <p>
      {{
        t(
          '注意：当前选择的是小额绿通，用于突发少量资源申领、紧急申领的场景，无需提前报需求预测。每单申领 100 核心以下时不需要审批，大于 100 核心时需要业务 leader 及资源团队审批。',
        )
      }}
    </p>
    <p style="color: red">1.小额绿通申领仅支持标准型（S类）16核心及以下云主机</p>
    <p style="color: red">
      2.小额绿通的机型为模糊匹配交付，在公司库存不足时，会自动切换为机型族内的其他机型，如需按提单指定机型交付，请勿使用此需求类型申领
    </p>
    <p style="color: red">3.模糊机型范围包括：S4m、S5、S5t、S6、S6t、SA2、SA3、SA5t、SA5、SA6、SA9、S9</p>
    <p>
      {{ t('如有疑问，请咨询') }}
      <w-name name="ICR" :alias="t('ICR(IEG资源服务助手)')" />
      {{ t('确认。') }}
    </p>
  </bk-alert>
</template>
