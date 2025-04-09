<script setup lang="ts">
import Panel from '@/components/panel';
import ExpeditingBtn from '@/views/ziyanScr/components/ticket-audit/children/expediting-btn.vue';

import type { IItsmTicketAudit } from './itsm-ticket-audit.vue';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';

const props = defineProps<{
  ticketAuditDetail: IItsmTicketAudit;
}>();

const { t } = useI18n();

const statusMap: Record<IItsmTicketAudit['status'], string> = {
  FINISHED: t('审批完成'),
  RUNNING: t('审批中...'),
  TERMINATED: t('审批被终止'),
};

const processors = computed(() =>
  props.ticketAuditDetail?.current_steps?.[0]?.processors?.filter((processor) => processor),
);
const processorsAuth = computed(() => props.ticketAuditDetail?.current_steps?.[0]?.processors_auth);
const processorsWithBizAccess = computed(() =>
  processors.value?.filter((processor) => processorsAuth.value[processor]),
);
const processorsWithoutBizAccess = computed(() =>
  processors.value?.filter((processor) => !processorsAuth.value[processor]),
);
</script>

<template>
  <panel class="home">
    <span class="status" v-if="ticketAuditDetail">
      <!-- icon -->
      <bk-loading
        v-if="ticketAuditDetail.status === 'RUNNING'"
        style="transform: scale(0.5)"
        mode="spin"
        theme="primary"
        loading
      ></bk-loading>
      <i v-else-if="ticketAuditDetail.status === 'TERMINATED'" class="hcm-icon bkhcm-icon-close-circle-fill"></i>
      <i v-else class="hcm-icon bkhcm-icon-7chenggong-01"></i>
      <!-- status -->
      <span>{{ statusMap[ticketAuditDetail.status] }}</span>
      <template v-if="ticketAuditDetail.status === 'RUNNING'">
        <div class="flex-row align-items-center">
          <span class="audit-status">
            {{ t('当前处于') }}
            <bk-tag theme="success" class="ml4 mr4">{{ ticketAuditDetail.current_steps?.[0]?.name }}</bk-tag>
            {{ t('环节') }}
          </span>
          <expediting-btn
            :processors="processors"
            :processors-with-biz-access="processorsWithBizAccess"
            :processors-without-biz-access="processorsWithoutBizAccess"
            :copy-text="t('复制 ITSM 审批单')"
            :ticket-link="ticketAuditDetail.itsm_ticket_link"
          />
        </div>
      </template>
      <!-- todo: error message 后端未提供 -->
    </span>
    <div v-else style="height: 22px"></div>
  </panel>
</template>

<style scoped lang="scss">
.home {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 52px;

  .status {
    display: flex;
    align-items: center;
    color: #313238;
    font-size: 14px;

    .audit-status {
      display: inline-flex;
      align-items: center;
      margin-left: 50px;
    }

    :deep(.hcm-icon) {
      font-size: 21px;
      margin-right: 13.5px;
      color: #3a84ff;
    }

    :deep(.bkhcm-icon-7chenggong-01) {
      color: #2dcb56;
    }

    :deep(.bkhcm-icon-close-circle-fill) {
      color: #cc4053;
    }
  }
}
</style>
