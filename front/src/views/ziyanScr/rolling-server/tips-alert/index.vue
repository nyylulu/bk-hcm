<script setup lang="ts">
import { Alert } from 'bkui-vue';
import WName from '@/components/w-name';
import { useI18n } from 'vue-i18n';

defineOptions({ name: 'RollingServerTipsAlert' });

const { t } = useI18n();
</script>

<template>
  <Alert>
    {{ t('注意：当前选择的是滚服项目，每个业务每月只允许申请使用2万核心，额度有限，限量申请。') }}
    <br />
    {{
      t(
        '若业务30天没有退回滚服项目资源，从超期第1天（按从申领资源算起，此时第31天）开始计算罚金成本，按超期机器的50%成本核算加收罚金，最少收1天，最多收91天。',
      )
    }}
    <br />
    {{ t('对于超期设备退回，当业务下有退回设备，计入滚服退回（非走滚服项目），罚金终止。') }}
    <br />
    {{
      t(
        '例：某业务申请2W核心滚服资源，有1.5W核心在30天内退还（此时无罚金），另外0.5W核心没有在30天内退回，从超期第1天开始按天计算50%罚金成本，此后业务下退回开始按天计算退回核心数（只计入CVM通用资源），直至退回核心数满0.5W核心，停止罚金成本。',
      )
    }}
    <br />
    <br />
    {{ t('如有疑问请咨询') }}
    <WName name="ICR" :alias="t('ICR(IEG资源服务助手)')" />
    {{ t('确认。') }}
  </Alert>
</template>

<style scoped></style>
