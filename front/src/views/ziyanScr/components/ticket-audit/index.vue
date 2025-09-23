<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { ITimelineItem } from './typings';

import { Copy, Share } from 'bkui-vue/lib/icon';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

interface IProp {
  title?: string;
  loading?: boolean;
  ticketLink?: string;
  logs: ITimelineItem[];
  copyText?: string;
}

// 审批流信息通用模板
defineOptions({ name: 'ticket-audit' });
defineProps<IProp>();

const { t } = useI18n();
</script>

<template>
  <div>
    <!-- loading态 -->
    <bk-loading loading v-if="loading">
      <div style="width: 100%; height: 300px" />
    </bk-loading>
    <!-- 审批流信息展示 -->
    <div v-else class="ticket-audit-wrapper">
      <div class="header">
        <!-- title -->
        <slot name="title">
          <div class="title">{{ title }}</div>
        </slot>
        <!-- tools slot -->
        <slot name="tools"></slot>
        <!-- copy -->
        <copy-to-clipboard v-if="ticketLink" class="ml12" :content="ticketLink">
          <bk-button theme="primary" text>
            <copy width="18" height="18" />
            {{ copyText }}
          </bk-button>
        </copy-to-clipboard>
        <slot name="toolkit"></slot>
        <!-- link -->
        <bk-link
          v-if="ticketLink"
          class="link-wrap"
          theme="primary"
          target="_blank"
          :disabled="!ticketLink"
          :href="ticketLink"
        >
          <div class="link-content">
            <span>{{ t('单据详情') }}</span>
            <share width="14" height="14" class="ml6" />
          </div>
        </bk-link>
      </div>
      <div :class="{ content: logs.length }">
        <bk-timeline :list="logs" />
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.ticket-audit-wrapper {
  .header {
    display: flex;
    align-items: center;
    margin-bottom: 12px;

    .title {
      color: $font-deep-color;
      font-weight: 700;
    }

    .link-wrap {
      margin-left: auto;

      .link-content {
        display: flex;
        align-items: center;
      }
    }
  }

  .content {
    padding: 12px 24px 0;
    background: #f5f7fa;

    :deep(.bk-timeline) {
      .bk-timeline-dot {
        padding-bottom: 0;

        .bk-timeline-content {
          max-width: 100%;
        }
      }
    }
  }
}
</style>
