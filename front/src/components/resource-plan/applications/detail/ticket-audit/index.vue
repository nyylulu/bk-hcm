<script setup lang="ts">
import { h, PropType, ref, VNode, watch } from 'vue';
import type {
  IPlanTicketAudit,
  IPlanTicketAuditLog,
  IPlanTicketCrpAudit,
  IPlanTicketItsmAudit,
} from '@/typings/resourcePlan';
import { useI18n } from 'vue-i18n';

import { WeixinPro } from 'bkui-vue/lib/icon';
import Panel from '@/components/panel';
import Step from './step.vue';
import WName from '@/components/w-name';

interface Step {
  name: string;
  auto?: boolean;
  link?: VNode;
}

defineOptions({ name: 'resource-plan-ticket-audit' });
const props = defineProps({ detail: Object as PropType<IPlanTicketAudit> });

const { t } = useI18n();

const steps = {
  itsm: [
    { name: t('提单'), auto: true },
    { name: t('直接上级审批') },
    { name: t('资源管理员') },
    { name: t('CRP系统审核') },
    { name: t('流程结束'), auto: true },
  ] as Step[],
  crp: [
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
  ] as Step[],
};

const renderItsmAuditLogs = ref([]);
const renderCrpAuditLogs = ref([]);

const renderLogs = (audit: IPlanTicketItsmAudit | IPlanTicketCrpAudit, type: 'itsm' | 'crp') => {
  const { current_steps, status, status_name, logs } = audit;

  // itsm没有提供name字段，直接展示message；crp正常展示name字段
  const tagKey = type === 'itsm' ? 'message' : 'name';

  // 审批通过的content展示：itsm只展示操作时间；crp展示操作人、信息、操作时间
  const renderSuccessContentValue = (log: IPlanTicketAuditLog) => {
    if (!log) return '';
    return type === 'itsm'
      ? h('p', { class: 'content' }, log.operate_at)
      : h('div', { class: 'content' }, [
          h('p', { class: 'message' }, [h(WName, { name: log.operator, class: 'mr4' }), `${log.message}`]),
          h('p', { class: 'time' }, log.operate_at),
        ]);
  };

  // 撤销、失败、结束、拒绝，done 这五种状态，直接输出 logs
  if (['revoked', 'rejected', 'canceled', 'failed', 'done'].includes(status)) {
    if (!logs) return [];
    const result = [
      ...logs.map((log) => ({ tag: h('div', { class: 'tag' }, log[tagKey]), content: renderSuccessContentValue(log) })),
      ...current_steps.map((step) => ({
        tag: h('div', { class: 'tag' }, step.name),
        content: h('div', { class: 'content' }, [
          `${status_name}，请联系 `,
          step.processors.map((processor, index) => {
            if (index < step.processors.length - 1) return [h(WName, { name: processor }), ', '];
            return h(WName, { name: processor });
          }),
        ]),
      })),
    ];
    return result.map((log, index) => ({
      ...log,
      nodeType: 'vnode',
      type: index < result.length - 1 || status === 'done' ? 'success' : 'danger',
    }));
  }

  // auditing 状态展示全部流程（current_steps可能不存在）
  const currentStepIdx = steps[type].findIndex((step) => step.name === current_steps?.[0]?.name);
  return steps[type].map((step, index) => {
    const { name, auto, link } = step;

    // 已审批通过的节点
    if (index < currentStepIdx || currentStepIdx === -1) {
      return {
        tag: h('div', { class: 'tag' }, [logs[index]?.[tagKey] ?? name, status === 'auditing' ? link : null]),
        content: auto && !logs[index] ? '' : renderSuccessContentValue(logs[index]),
        nodeType: 'vnode',
        // 单据为itsm时，如果itsm审批通过，crp审批未通过，并且logs中没有记录时，显示为danger状态；否则，显示success状态
        type:
          type === 'itsm' && audit.status === 'done' && props.detail.crp_audit.status !== 'done' && !logs[index]
            ? 'danger'
            : 'success',
      };
    }

    // 当前节点
    if (index === currentStepIdx) {
      const { name, processors } = current_steps[0];
      return {
        tag: h('div', { class: 'tag current' }, name),
        content: h('div', { class: 'content' }, [
          processors.map((processor, index) => {
            if (index < processors.length - 1) return [h(WName, { name: processor }), ', '];
            return h(WName, { name: processor, class: 'mr4' });
          }),
          t('正在审批中...'),
        ]),
        nodeType: 'vnode',
        icon: h('i', { class: 'hcm-icon bkhcm-icon-jiazai' }),
        type: 'primary',
      };
    }

    // 未来节点
    return { tag: h('div', { class: 'tag' }, logs[index]?.[tagKey] ?? name), content: '', nodeType: 'vnode' };
  });
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
    renderItsmAuditLogs.value = renderLogs(itsm_audit, 'itsm');
    if (itsm_audit?.status === 'done') renderCrpAuditLogs.value = renderLogs(crp_audit, 'crp');
    // 如果itsm处于审批中状态，crp审批信息展示静态数据
    else if (itsm_audit?.status === 'auditing')
      renderCrpAuditLogs.value = steps.crp.map((step) => ({
        tag: h('div', { class: 'tag' }, step.name),
        content: '',
        nodeType: 'vnode',
      }));
  },
  { immediate: true, deep: true },
);
</script>

<template>
  <panel v-if="renderItsmAuditLogs.length" class="panel" :title="t('审批信息')">
    <step :label="t('一级审批')" :title="t('ITSM 平台审批')" :link-url="detail?.itsm_audit?.itsm_url">
      <template #content>
        <bk-timeline :list="renderItsmAuditLogs" />
      </template>
    </step>
    <step
      v-if="renderCrpAuditLogs.length"
      :label="t('二级审批')"
      :title="t('CRP 平台审批')"
      :link-url="detail?.crp_audit?.crp_url"
    >
      <template #content>
        <bk-timeline :list="renderCrpAuditLogs" />
      </template>
    </step>
  </panel>
</template>

<style scoped lang="scss">
.panel {
  margin-bottom: 16px;

  :deep(.bk-timeline) {
    .tag {
      display: flex;
      align-items: center;
      color: #313238;
      line-height: 22px;
      &.current {
        font-weight: 700;
        font-size: 16px;
        color: #313238;
        line-height: 24px;
      }
    }

    .content {
      font-size: 12px;
      line-height: 20px;
      .message {
        color: #4d4f56;
      }
      .time {
        color: #979ba5;
      }
    }

    .bk-timeline-dot {
      padding-bottom: 0;

      .bk-timeline-content {
        max-width: 100%;
      }
    }
  }
}
</style>
