<script setup lang="ts"></script>

<template>
  <div>
    <bk-alert closable class="info-alert">
      <template #title>全平台额度，是按自然月统计的当月数据。下月数据会自动计算。</template>
    </bk-alert>
    <div class="info-grid">
      <div class="row">
        <div class="cell head">总限额 (全平台)</div>
        <div class="cell">20,000 核</div>
      </div>
      <div class="row">
        <div class="cell head">基础额度 (单业务)</div>
        <div class="cell">20,000 核</div>
      </div>
      <div class="row">
        <div class="cell head">剩余额度 (全平台)</div>
        <div class="cell">0 核</div>
      </div>
      <div class="row">
        <div class="cell head">已交付 (全业务)</div>
        <div class="cell">200 核</div>
      </div>
      <div class="row">
        <div class="cell head">已退还 (全业务)</div>
        <div class="cell">20 核</div>
      </div>
      <div class="row refund">
        <div class="cell head">应退还 (资源池)</div>
        <div class="cell"><em class="hl">20 核</em></div>
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
    &.refund {
      .cell {
        .hl {
          font-weight: 700;
          color: #e38b02;
          font-style: normal;
        }
        &.head {
          background: #fdf4e8;
        }
      }
    }
  }
}
</style>
