<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import HcmLink from '@/components/hcm-link/index.vue';
import { RequirementType } from '@/store/config/requirement';

defineProps<{ requireType?: RequirementType }>();

defineOptions({ name: 'HostApplyTipsAlert' });

const { t } = useI18n();
</script>

<template>
  <!-- eslint-disable prettier/prettier -->
  <bk-alert class="host-apply-tips-alert" theme="warning">
    <p>{{ t('主机申请流程') }}</p>
    <ul>
      <ol>
        <span class="label">1.{{ t('报备资源预测：') }}</span>
        <span class="content" v-if="requireType !== RequirementType.Dissolve">
          {{ t('业务提前13周提交13周后日期的预测需求，可按包年包月申请。13周内提资源预测，公司资源团队不保证主机能够及时交付，计费模式为按量计费，成本高，参考') }}
          <hcm-link theme="primary" size="small" href="https://iwiki.woa.com/p/4011431167" target="_blank">{{ t('计费模式说明') }}</hcm-link>
        </span>
        <span v-else>
          {{ t('业务需提报预测需求，允许在13周内提报，计费模式可根据需求选包年包月、按量计费') }}
        </span>
      </ol>
      <ol>
        <span class="label">2.{{ t('申领主机：') }}</span>
        <span class="content">{{ t('主机申领关联资源预测需求，无预测时将无法提单申领主机') }}</span>
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
