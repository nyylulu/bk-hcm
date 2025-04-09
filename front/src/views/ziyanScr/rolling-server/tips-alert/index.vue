<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { Alert } from 'bkui-vue';
import WName from '@/components/w-name';

defineOptions({ name: 'RollingServerTipsAlert' });

const { t } = useI18n();
</script>

<template>
  <!-- eslint-disable prettier/prettier -->
  <Alert class="rolling-server-tips-alert" theme="warning">
    <p>{{ t('注意：当前选择的是滚服项目，每个业务每月只允许申请使用1万核心，额度有限，限量申请，先到先得。更多信息请参考文档') }} <bk-link theme="primary" href="https://iwiki.woa.com/p/4012608772" target="_blank">https://iwiki.woa.com/p/4012608772</bk-link></p>
    <ul>
      <li><strong>{{ t('套餐计费：') }}</strong>{{ t('继承原有机器的计费起止时间、计费模式') }}</li>
      <li><strong>{{ t('机型限制：') }}</strong>{{ t('只支持通用机型（通用/专用机型，请参考公司定义说明') }} <bk-link theme="primary" href="https://doc.weixin.qq.com/sheet/e3_ACAAPwaDACcY182EYUTTh6Fw0xsO9?scode=AJEAIQdfAAooKA1w6rAJwAVAbdAFw&tab=BB08J2" target="_blank">https://doc.weixin.qq.com/sheet/e3_ACAAPwaDACcY182EYUTTh6Fw0xsO9?scode=AJEAIQdfAAooKA1w6rAJwAVAbdAFw&tab=BB08J2</bk-link>{{ t('）') }}</li>
      <li><strong>{{ t('地域限制：') }}</strong>{{ t('支持同城内所有可用区继承，如继承机器是上海的机器，可提交所有上海可用区机器申领') }}</li>
      <li><strong>{{ t('退回要求：') }}</strong>{{ t('应在30天内退回等量的相应要求的机器，超期会产生罚金。注意，处于机房裁撤的机器，回收时不会匹配滚服的应退回额度。') }}</li>
      <li><strong>{{ t('超期罚金：') }}</strong>{{ t('在机器申请后第31-121天内未退还，按超期机器的50%成本核算加收罚金，最少收1天，最多收91天') }}</li>
    </ul>
    <p>{{ t('例：某业务申请1W核心滚服资源，有0.5W核心在30天内退还（此时无罚金），另外0.5W核心没有在30天内退回，从超期第1天开始按天计算50%罚金成本，此后业务下退回开始按天计算退回核心数（只计入CVM通用资源），直至退回核心数满0.5W核心，停止罚金成本。') }}</p>
    <br />
    <p>{{ t('如有疑问请咨询') }}<WName name="ICR" :alias="t('ICR(IEG资源服务助手)')" />{{ t('确认。') }}</p>
  </Alert>
</template>

<style scoped lang="scss">
.rolling-server-tips-alert {
  .bk-link {
    font-size: 12px;
    word-break: break-all;
  }
}
</style>
