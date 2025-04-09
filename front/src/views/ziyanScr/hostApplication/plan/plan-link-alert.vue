<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

withDefaults(defineProps<{ bkBizId: number; showSuggestions?: boolean }>(), {
  showSuggestions: true,
});

const { t } = useI18n();
</script>

<template>
  <!-- 主机申领 & CVM生产 - 预测指引 -->
  <bk-alert theme="warning" class="plan-link-alert">
    {{ t('该地域，在当月，没有可申领的预测需求') }}
    <template v-if="showSuggestions">
      {{ t('，建议：') }}
      <ul>
        <li>
          <span>1.{{ t('切换至有预测需求的地域，') }}</span>
          <bk-button
            theme="primary"
            text
            @click="routerAction.open({ path: '/business/resource-plan', query: { [GLOBAL_BIZS_KEY]: bkBizId } })"
          >
            {{ t('查询当前预测需求') }}
          </bk-button>
        </li>
        <li>
          <span>2.{{ t('请先提交预测单，将期望到货日期设置为当月，预测需求审批通过后可申领主机，') }}</span>
          <bk-button
            theme="primary"
            text
            @click="routerAction.open({ path: '/business/resource-plan/add', query: { [GLOBAL_BIZS_KEY]: bkBizId } })"
          >
            {{ t('去新增资源预测') }}
          </bk-button>
        </li>
      </ul>
    </template>
  </bk-alert>
</template>

<style scoped lang="scss">
.plan-link-alert {
  line-height: normal;
}
</style>
