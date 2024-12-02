<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { ITimelineItem } from './typings';

import { Share } from 'bkui-vue/lib/icon';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

interface IProp {
  title: string;
  loading: boolean;
  ticketLink: string;
  logs: ITimelineItem[];
}

// 审批流信息通用模板
defineOptions({ name: 'ticket-audit' });
defineProps<IProp>();

const { t } = useI18n();
</script>

<template>
  <bk-loading :loading="loading" class="ticket-audit-wrapper">
    <div class="header">
      <div class="title">{{ title }}</div>
      <bk-link theme="primary" target="_blank" :disabled="!ticketLink" :href="ticketLink">
        <div class="link-content">
          <span>{{ t('单据详情') }}</span>
          <Share width="14" height="14" class="ml6" />
        </div>
      </bk-link>
      <copy-to-clipboard class="ml6" :disabled="!ticketLink" :content="ticketLink" />
      <slot name="header-end"></slot>
    </div>
    <div class="content">
      <bk-timeline :list="logs" />
    </div>
  </bk-loading>
</template>

<style scoped lang="scss">
.ticket-audit-wrapper {
  .header {
    display: flex;
    align-items: center;
    margin-bottom: 12px;

    .title {
      margin-right: 24px;
      color: $font-deep-color;
      font-weight: 700;
    }

    .link-content {
      display: flex;
      align-items: center;
    }
  }

  .content {
    display: flow-root;
    padding: 0 100px;

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
