<script setup lang="ts">
import { h, onBeforeMount, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { useZiyanScrStore } from '@/store';
import type {
  IApplyCrpTicketAudit,
  IApplyCrpTicketAuditCurrentStepItem,
  IApplyCrpTicketAuditFailInfoItem,
  IApplyCrpTicketAuditLogItem,
} from '@/typings/ziyanScr';
import {
  ITimelineIconStatusType,
  ITimelineNodeType,
  type ITimelineItem,
} from '@/views/ziyanScr/components/ticket-audit/typings';

import { OverflowTitle } from 'bkui-vue';
import TicketAudit from '@/views/ziyanScr/components/ticket-audit/index.vue';
import TimelineTag from '@/views/ziyanScr/components/ticket-audit/children/timeline-tag.vue';
import TimelineContent from '@/views/ziyanScr/components/ticket-audit/children/timeline-content.vue';
import LoadingIcon from '@/views/ziyanScr/components/ticket-audit/icon/loading-icon.vue';
import { getWNameVNodeList } from '@/views/ziyanScr/components/ticket-audit/utils';

const props = defineProps<{ crpTicketId: string; subOrderId: string }>();

const ziyanScrStore = useZiyanScrStore();
const { t } = useI18n();

const loading = ref(false);
const getDefaultData = (): Partial<IApplyCrpTicketAudit> => ({ crp_ticket_link: '', logs: [], current_step: null });
const data = reactive(getDefaultData());
const renderLogs = ref<ITimelineItem[]>([]);

const steps = [
  t('部门管理员审批'),
  t('业务总监审批'),
  t('规划经理审批'),
  t('资源经理审批'),
  t('等待云上审批'),
  t('等待交付'),
  t('交付队列中'),
  t('流程结束'),
];

const getData = async (crp_ticket_id: string, suborder_id: string) => {
  if (!crp_ticket_id) {
    renderLogs.value = steps.map((item) => getDefaultLog(item));
    return;
  }

  loading.value = true;
  try {
    const res = await ziyanScrStore.getApplyCrpTicketAudit({ crp_ticket_id, suborder_id });
    Object.assign(data, res);

    // 构建timeline节点
    renderLogs.value = getRenderLogs(res);
  } catch (error) {
    console.error(error);
    Object.assign(data, getDefaultData());
  } finally {
    loading.value = false;
  }
};

// 同意节点
const getSuccessLog = (log: IApplyCrpTicketAuditLogItem): ITimelineItem => {
  const { task_name, operate_result, operator, operate_time } = log;
  const operators = operator.split(';');

  return {
    tag: h(TimelineTag, null, task_name),
    content: h(TimelineContent, null, [
      h('span', null, `${t('审批意见：')}${operate_result}`),
      h('span', null, [t('审批人：'), getWNameVNodeList(operators)]),
      h('span', null, `${t('审批时间：')}${operate_time}`),
    ]),
    nodeType: ITimelineNodeType.VNode,
    type: 'success',
  };
};
// 失败节点
const getFailLog = (log: IApplyCrpTicketAuditLogItem, failInfo: IApplyCrpTicketAuditFailInfoItem): ITimelineItem => {
  const { task_name } = log;
  const { error_type, error_msg, error_msg_type_cn, operator } = failInfo;
  const operators = operator.split(';');

  // 兼容空值
  const errorMsgTypeCn = error_msg_type_cn ? `, ${error_msg_type_cn}` : '';
  const errorMsg = error_msg ? `, ${error_msg}` : '';

  return {
    tag: h(TimelineTag, null, task_name),
    content: h(
      TimelineContent,
      null,
      h('span', { class: 'error-message' }, [
        t('错误信息：'),
        h(OverflowTitle, { type: 'tips', class: 'error-message-content' }, `${error_type}${errorMsgTypeCn}${errorMsg}`),
        t('。'),
        h('span', null, [t('如有疑问，请联系：'), getWNameVNodeList(operators)]),
      ]),
    ),
    nodeType: ITimelineNodeType.VNode,
    type: 'danger',
  };
};
// 审批中节点
const getAuditingLog = (current_step: IApplyCrpTicketAuditCurrentStepItem): ITimelineItem => {
  const { current_task_name, status, status_desc } = current_step;

  return {
    tag: h(TimelineTag, { isCurrent: true }, current_task_name),
    content: h(TimelineContent, null, [
      h('span', null, `${t('单据状态码：')}${status}`),
      h('span', null, `${t('单据状态描述：')}${status_desc}`),
    ]),
    nodeType: ITimelineNodeType.VNode,
    icon: h(LoadingIcon),
  };
};
// 默认节点
const getDefaultLog = (name: string, type: ITimelineIconStatusType = 'default'): ITimelineItem => {
  return { tag: h(TimelineTag, null, name), nodeType: ITimelineNodeType.VNode, type };
};
const getRenderLogs = ({ logs, current_step }: IApplyCrpTicketAudit) => {
  let renderLogs: ITimelineItem[];

  const { current_task_no, fail_instance_info, status } = current_step;

  // 结束态：正常通过、失败、驳回
  if (current_task_no === -1) {
    renderLogs = logs.map<ITimelineItem>((log, index) => {
      const { operate_result, task_name } = log;

      // 同意节点
      if ('同意' === operate_result) return getSuccessLog(log);

      // todo：目前CRP方无法提供 '驳回' === operate_result 的场景，所以暂时不处理驳回log的情况；驳回处理放置到 fail 情况下特殊处理
      if (index === logs.length - 1 && status === 127) {
        const whitelist = ['cutechen', 'lotuschen'];
        return {
          tag: h(TimelineTag, null, task_name),
          // !! 需要加白处理
          content: h(
            TimelineContent,
            { class: 'no-gap' },
            h('div', null, [t('单据已被驳回，如有疑问，请联系：'), getWNameVNodeList(whitelist)]),
          ),
          nodeType: ITimelineNodeType.VNode,
          type: 'danger',
        };
      }

      // 失败节点：只展示最后一个错误信息
      if (index === logs.length - 1 && fail_instance_info !== null) {
        return getFailLog(log, fail_instance_info[fail_instance_info.length - 1]);
      }

      // 默认处理
      return getDefaultLog(task_name, 'success');
    });
  }
  // 审批中
  else {
    renderLogs = logs.map<ITimelineItem>((log) => {
      const { task_no, task_name } = log;

      // 已审批节点
      if (task_no < current_task_no) return getSuccessLog(log);

      // 审批中节点
      if (task_no === current_task_no) return getAuditingLog(current_step);

      // 未来节点
      return getDefaultLog(task_name);
    });
  }

  return renderLogs;
};

onBeforeMount(() => {
  getData(props.crpTicketId, props.subOrderId);
});
</script>

<template>
  <ticket-audit
    class="crp-ticket-audit"
    :title="t('CRP平台审批')"
    :loading="loading"
    :ticket-link="data.crp_ticket_link"
    :logs="renderLogs"
    :copy-text="t('复制CRP审批单')"
  />
</template>

<style scoped lang="scss">
.crp-ticket-audit {
  padding: 16px;

  .cancel-btn {
    margin-left: auto;
    min-width: 88px;
  }

  :deep(.i-timeline-content) {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 12px;

    &.no-gap {
      gap: 0;
    }

    .error-message {
      display: inline-flex;
      align-items: center;
      color: $danger-color;

      .error-message-content {
        max-width: 300px;
      }
    }
  }
}
</style>
