<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { useGreenChannelQuotaStore } from '@/store/green-channel/quota';

const greenChannelQuotaStore = useGreenChannelQuotaStore();
const { globalQuotaConfig } = storeToRefs(greenChannelQuotaStore);

const formatValue = (value: number) => {
  if (isNaN(value) || value === undefined) {
    return '--';
  }
  return value.toLocaleString();
};
</script>

<template>
  <div>
    <bk-alert closable class="info-alert">
      <template #title>全平台额度，是按自然月统计的当月数据。下月数据会自动计算。</template>
    </bk-alert>
    <div class="info-grid">
      <div class="row">
        <div class="cell head">总限额 (全平台)</div>
        <div class="cell">{{ formatValue(globalQuotaConfig.ieg_quota) }} 核</div>
      </div>
      <div class="row">
        <div class="cell head">基础额度 (单业务/周)</div>
        <div class="cell">{{ formatValue(globalQuotaConfig.biz_quota) }} 核</div>
      </div>
      <div class="row">
        <div class="cell head">需要人工审批核数/单</div>
        <div class="cell">{{ formatValue(globalQuotaConfig.audit_threshold) }} 核</div>
      </div>
      <div class="row">
        <div class="cell head">已交付 (全业务)</div>
        <div class="cell">{{ formatValue(globalQuotaConfig.sum_delivered_core) }} 核</div>
      </div>
      <div class="row">
        <div class="cell head">剩余额度 (全平台)</div>
        <div class="cell">{{ formatValue(globalQuotaConfig.ieg_quota - globalQuotaConfig.sum_delivered_core) }} 核</div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.info-alert {
  margin-bottom: 20px;
}
.info-grid {
  display: grid;
  grid-template-columns: 1fr;
  grid-gap: 0;
  .row {
    display: grid;
    grid-template-columns: 260px 1fr;
    gap: 0;
    .cell {
      font-size: 12px;
      color: #313238;
      padding: 12px 16px;
      border: 1px solid #dcdee5;
      margin-left: -1px;
      margin-top: -1px;
      &.head {
        background: #fafbfd;
      }
    }
  }
}
</style>
