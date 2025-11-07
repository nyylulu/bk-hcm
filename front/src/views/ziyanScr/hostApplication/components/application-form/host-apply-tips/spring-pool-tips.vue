<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <bk-alert class="host-apply-spring-pool-alert" theme="warning">
    <p>{{ t('注意：') }}</p>
    <ul>
      <ol>
        <span class="content">
          {{ t('1、“春保资源池”资源申请，需先联系“ICR(IEG资源服务助手)”沟通需求实际情况，授权额度后可进行资源申领') }}
        </span>
      </ol>
      <ol>
        <span class="content">{{ t('2、资源总额度有限，先到先得，按量计费（170%）') }}</span>
      </ol>
      <ol>
        <span class="content">{{ t('3、资源相关所有操作及成本，遵循公司“常规项目”相关规则') }}</span>
      </ol>
    </ul>
  </bk-alert>
</template>

<style scoped lang="scss">
.host-apply-spring-pool-alert {
  .label {
    font-weight: 700;
  }
}
</style>
