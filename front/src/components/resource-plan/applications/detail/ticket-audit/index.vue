<script setup lang="ts">
import { h, onUnmounted, PropType, ref, VNode, watch } from 'vue';
import type {
  IPlanTicketAudit,
  IPlanTicketAuditLog,
  IPlanTicketCrpAudit,
  IPlanTicketItsmAudit,
} from '@/typings/resourcePlan';
import { useI18n } from 'vue-i18n';
import { useUserStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { TimeoutPollAction } from '@/hooks/use-timeout-poll';
import {
  type ITimelineIconStatusType,
  type ITimelineItem,
  ITimelineNodeType,
} from '@/views/ziyanScr/components/ticket-audit/typings';
import http from '@/http';

import { Message } from 'bkui-vue';
import { WeixinPro, ExclamationCircleShape } from 'bkui-vue/lib/icon';
import Panel from '@/components/panel';
import WName from '@/components/w-name';
import TicketAudit from '@/views/ziyanScr/components/ticket-audit/index.vue';
import TimelineTag from '@/views/ziyanScr/components/ticket-audit/children/timeline-tag.vue';
import TimelineContent from '@/views/ziyanScr/components/ticket-audit/children/timeline-content.vue';
import LoadingIcon from '@/views/ziyanScr/components/ticket-audit/icon/loading-icon.vue';
import ApprovalBtn from '@/views/ziyanScr/components/ticket-audit/children/approval-btn.vue';
import { getWNameVNodeList } from '@/views/ziyanScr/components/ticket-audit/utils';

interface Step {
  name: string;
  auto?: boolean;
  link?: VNode;
}

defineOptions({ name: 'resource-plan-ticket-audit' });
const props = defineProps({
  detail: Object as PropType<IPlanTicketAudit>,
  fetchData: Function as PropType<() => Promise<void>>,
  timeoutPollAction: Object as PropType<TimeoutPollAction>,
  isBusinessPage: Boolean,
});

const { t } = useI18n();
const { getBusinessApiPath } = useWhereAmI();
const userStore = useUserStore();

const crpSteps: Step[] = [
  {
    name: t('部门管理员审批'),
    link: h(
      WName,
      { name: 'ICR', alias: t('企微催单（ICR助手）'), class: 'ml8' },
      { icon: () => h(WeixinPro, { width: 16, height: 16, class: 'mr5' }) },
    ),
  },
  { name: t('规划经理审批') },
  { name: t('需求经理审批') },
  { name: t('架平审批'), auto: true },
  { name: t('资源总监审批'), auto: true },
  { name: t('流程结束'), auto: true },
];

const renderItsmAuditLogs = ref<ITimelineItem[]>([]);
const renderCrpAuditLogs = ref<ITimelineItem[]>([]);

// 历史节点的tag展示：itsm没有提供name字段，直接展示message；crp正常展示name字段
const getHistoryStepTag = (log: IPlanTicketAuditLog, auditType: 'itsm' | 'crp') => {
  const key = auditType === 'itsm' ? 'message' : 'name';
  return h(TimelineTag, null, log[key]);
};
// 历史节点的content展示：itsm只展示操作时间；crp展示操作人、信息、操作时间
const getHistoryStepContent = ({ operate_at, operator, message }: IPlanTicketAuditLog, auditType: 'itsm' | 'crp') => {
  return auditType === 'itsm'
    ? h(TimelineContent, { class: 'time-value' }, operate_at)
    : h(TimelineContent, null, [
        h('p', { class: 'message' }, [h(WName, { name: operator, class: 'mr4' }), `${message}`]),
        h('p', { class: 'time-value' }, operate_at),
      ]);
};

const getHistoryStepItems = (audit: IPlanTicketItsmAudit | IPlanTicketCrpAudit, auditType: 'itsm' | 'crp') => {
  const { logs, status } = audit;

  // 兼容 crp init/failed 的情况
  if (!logs) return [];

  return logs.map<ITimelineItem>((log, index) => {
    // 如果是最后一个节点，且单据状态为revoked、canceled、rejected，则展示为danger，否则为success
    const type: ITimelineIconStatusType =
      ['revoked', 'canceled', 'rejected'].includes(status) && index === logs.length - 1 ? 'danger' : 'success';

    return {
      tag: getHistoryStepTag(log, auditType),
      content: getHistoryStepContent(log, auditType),
      nodeType: ITimelineNodeType.VNode,
      type,
    };
  });
};

// auditing展示审批操作、failed展示审批失败原因
const getCurrentStepItems = (
  { current_steps, status, status_name }: IPlanTicketItsmAudit | IPlanTicketCrpAudit,
  showApproval = true,
) => {
  // 兼容 crp init/failed 的情况
  if (!current_steps) return [];

  return current_steps.map<ITimelineItem>(({ name, processors }) => {
    const isAuditing = status === 'auditing';
    // auditing状态，且用户为审批人时，可以进行审批处理
    const hasApprovalBtn = showApproval && isAuditing && processors.includes(userStore.username);

    return {
      tag: h(TimelineTag, { isCurrent: true }, [
        name,
        hasApprovalBtn
          ? h(
              ApprovalBtn,
              {
                class: 'ml24',
                loading: approvalLoading.value,
                confirmHandler: approvalItsmAudit,
                onShown: () => {
                  // 当进入审批操作的时候，暂停定时刷新任务
                  props.timeoutPollAction?.reset();
                },
                onHidden: () => {
                  // 当退出审批操作的时候，恢复定时刷新任务
                  props.timeoutPollAction?.resume();
                },
              },
              t('立即处理'),
            )
          : null,
      ]),
      content: h(TimelineContent, null, [`${status_name}，请联系 `, getWNameVNodeList(processors)]),
      nodeType: ITimelineNodeType.VNode,
      icon: isAuditing ? h(LoadingIcon) : undefined,
      type: isAuditing ? 'primary' : 'danger',
    };
  });
};

const renderItsmLogs = (audit: IPlanTicketItsmAudit) => {
  return [...getHistoryStepItems(audit, 'itsm'), ...getCurrentStepItems(audit)];
};

const getStaticCrpLogs = () =>
  crpSteps.map<ITimelineItem>(({ name }) => ({ tag: h(TimelineTag, null, name), nodeType: ITimelineNodeType.VNode }));
const renderCrpLogs = (audit: IPlanTicketCrpAudit) => {
  // 空值兼容
  if (!audit?.crp_sn) return getStaticCrpLogs();

  const { status, logs, current_steps } = audit;

  // 撤销、失败、结束、拒绝，done 这五种状态，直接输出response
  if (['revoked', 'rejected', 'canceled', 'failed', 'done'].includes(status)) {
    return [...getHistoryStepItems(audit, 'crp'), ...getCurrentStepItems(audit, false)]; // CRP暂不支持审批
  }

  const currentStepIdx = crpSteps.findIndex((step) => step.name === current_steps?.[0]?.name);

  // auditing 状态展示全部流程
  return crpSteps.map<ITimelineItem>(({ name, auto, link }, index) => {
    // 已审批通过的节点
    if (index < currentStepIdx) {
      return {
        tag: h(TimelineTag, null, [logs[index]?.name ?? name, status === 'auditing' ? link : null]),
        content: auto && !logs[index] ? '' : getHistoryStepContent(logs[index], 'crp'),
        nodeType: ITimelineNodeType.VNode,
        type: 'success',
      };
    }

    // 当前节点
    if (index === currentStepIdx) return getCurrentStepItems(audit)[0];

    // 未来节点
    return { tag: h(TimelineTag, null, logs[index]?.name ?? name), content: '', nodeType: ITimelineNodeType.VNode };
  });
};

// 审批操作
const approvalLoading = ref(false);
let approvalLoadingTimer: ReturnType<typeof setTimeout> | null = null;
const approvalItsmAudit = async ({ approval, remark }: { approval: boolean; remark: string }) => {
  const { ticket_id } = props.detail;
  const { state_id } = props.detail.itsm_audit.current_steps[0];
  const params = { state_id, approval, remark };
  await http.post(`/api/v1/woa/${getBusinessApiPath()}plans/resources/tickets/${ticket_id}/approve_itsm_node`, params);

  Message({ theme: 'success', message: t('请求已提交，5s后自动刷新') });
  // 5s后刷新itsm审批流信息
  approvalLoading.value = true;
  renderItsmAuditLogs.value = renderItsmLogs(props.detail.itsm_audit); // 重新渲染timeline
  approvalLoadingTimer = setTimeout(() => {
    approvalLoading.value = false;
    props.fetchData();
  }, 5000);
};

watch(
  () => props.detail,
  (detail) => {
    if (!detail) {
      renderItsmAuditLogs.value = [];
      renderCrpAuditLogs.value = [];
      return;
    }

    const { itsm_audit, crp_audit } = detail;
    renderItsmAuditLogs.value = renderItsmLogs(itsm_audit);

    if (itsm_audit?.status === 'done') renderCrpAuditLogs.value = renderCrpLogs(crp_audit);
    // 如果itsm处于审批中状态，crp审批信息展示静态数据
    else if (itsm_audit?.status === 'auditing') renderCrpAuditLogs.value = getStaticCrpLogs();
  },
  { immediate: true, deep: true },
);

onUnmounted(() => {
  clearTimeout(approvalLoadingTimer);
  props.timeoutPollAction?.reset();
});
</script>

<template>
  <panel v-if="renderItsmAuditLogs.length" class="panel" :title="t('审批信息')">
    <div class="step-wrap">
      <h3 class="label">{{ t('一级审批') }}：</h3>
      <ticket-audit
        class="content"
        :title="t('ITSM 平台审批')"
        :ticket-link="detail?.itsm_audit?.itsm_url"
        :logs="renderItsmAuditLogs"
        :copy-text="t('复制ITSM审批单')"
      >
        <template #tools>
          <bk-popover
            v-if="
              isBusinessPage &&
              detail?.itsm_audit?.status === 'auditing' &&
              detail.itsm_audit.current_steps?.[0]?.processors.includes(userStore.username) &&
              !detail.itsm_audit.current_steps?.[0]?.processors_auth[userStore.username]
            "
            max-width="270"
            placement="top"
            :content="t('当前审批人无「业务访问」权限，请复制ITSM链接，在企业微信联系审批人进行催单')"
          >
            <exclamation-circle-shape class="ml12" width="18" height="18" fill="#EA3636" style="cursor: pointer" />
          </bk-popover>
        </template>
      </ticket-audit>
    </div>
    <div class="step-wrap" v-if="renderCrpAuditLogs.length">
      <h3 class="label">{{ t('二级审批') }}：</h3>
      <ticket-audit
        class="content"
        :title="t('CRP 平台审批')"
        :ticket-link="detail?.crp_audit?.crp_url"
        :logs="renderCrpAuditLogs"
        :copy-text="t('复制CRP审批单')"
      >
        <template #tools>
          <bk-popover
            v-if="
              isBusinessPage &&
              detail?.crp_audit?.status === 'auditing' &&
              detail.crp_audit.current_steps?.[0]?.processors.includes(userStore.username) &&
              !detail.crp_audit.current_steps?.[0]?.processors_auth[userStore.username]
            "
            max-width="270"
            placement="top"
            :content="t('当前审批人无「业务访问」权限，请复制CRP链接，在企业微信联系审批人进行催单')"
          >
            <exclamation-circle-shape class="ml12" width="18" height="18" fill="#EA3636" style="cursor: pointer" />
          </bk-popover>
        </template>
      </ticket-audit>
    </div>
  </panel>
</template>

<style scoped lang="scss">
.panel {
  margin-bottom: 16px;

  .step-wrap {
    display: flex;
    padding: 0 72px;

    .label {
      flex-shrink: 0;
      font-size: 14px;
      width: 80px;
      color: $danger-color;
    }

    .content {
      flex: 1;
    }

    &:not(:last-of-type) {
      margin-bottom: 16px;
    }
  }
}
</style>
