<script setup lang="ts">
import { h, onBeforeMount, reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { useUserStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useFormModel from '@/hooks/useFormModel';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import type { IQueryResData } from '@/typings';
import { type ITimelineItem, ITimelineNodeType } from '@/views/ziyanScr/components/ticket-audit/typings';
import http from '@/http';

import { Button, Message } from 'bkui-vue';
import TicketAudit from '@/views/ziyanScr/components/ticket-audit/index.vue';
import WName from '@/components/w-name';

interface ItsmTicketAudit {
  order_id: number;
  itsm_ticket_id: string;
  itsm_ticket_link: string;
  status: string;
  current_steps: {
    name: string;
    processors: string;
    state_id: number;
  }[];
  logs: {
    operator: string;
    operate_at: string;
    message: string;
    source: string;
  }[];
}

const props = defineProps<{ orderId: number; creator: string; bkBizId: number }>();

const userStore = useUserStore();
const { t } = useI18n();
const { getBusinessApiPath } = useWhereAmI();

const loading = ref(false);
const getDefaultData = (): Partial<ItsmTicketAudit> => ({ itsm_ticket_link: '', logs: [], current_steps: [] });
const data = reactive(getDefaultData());
const renderLogs = ref<ITimelineItem[]>([]);

const getData = async (order_id: number, bk_biz_id: number) => {
  loading.value = true;
  try {
    const res: IQueryResData<ItsmTicketAudit> = await http.post(
      `/api/v1/woa/${getBusinessApiPath()}task/get/apply/ticket/audit`,
      { order_id, bk_biz_id },
    );

    Object.assign(data, res.data);

    // 构建timeline节点
    renderLogs.value = getRenderLogs(res.data);

    // 如果单据处于处理中(RUNNING)状态, 创建定时任务(30s刷新一次, 最多刷新60次)
    if (data.status === 'RUNNING') {
      refreshTask.resume();
    }
  } catch (error) {
    console.error(error);
    Object.assign(data, getDefaultData());
  } finally {
    loading.value = false;
  }
};
const getRenderLogs = ({ logs, current_steps }: ItsmTicketAudit) => {
  // 本期直接输出 logs + current_steps
  return [
    ...logs.map<ITimelineItem>(({ message, operate_at }) => ({
      tag: h('div', { class: 'i-timeline-tag' }, message),
      content: h('div', { class: 'i-timeline-content' }, operate_at),
      nodeType: ITimelineNodeType.VNode,
      type: 'success',
    })),
    ...current_steps.map<ITimelineItem>(({ name, processors }) => {
      if (processors === '系统自动处理') {
        return {
          tag: h('div', { class: 'i-timeline-tag' }, name),
          content: h('div', { class: 'i-timeline-content current-step' }, processors),
          nodeType: ITimelineNodeType.VNode,
        };
      }

      const processorsArr = processors.split(',');
      // 当前用户为审批人时，可以进行审批处理
      const hasHandleBtn = processorsArr.includes(userStore.username);

      return {
        tag: h('div', { class: 'i-timeline-tag' }, [
          name,
          hasHandleBtn
            ? h(
                Button,
                { theme: 'primary', size: 'small', class: 'approval-btn', onClick: () => (isShow.value = true) },
                t('立即处理'),
              )
            : null,
        ]),
        content: h('div', { class: 'i-timeline-content current-step' }, [
          processorsArr.map((processor, index) => {
            if (index < processors.length - 1) return [h(WName, { name: processor }), ', '];
            return h(WName, { name: processor, class: 'mr4' });
          }),
          t('正在审批中...'),
        ]),
        nodeType: ITimelineNodeType.VNode,
      };
    }),
  ];
};

// 审批操作
const isShow = ref(false);
const { formModel, resetForm } = useFormModel({ approval: true, remark: '' });
const approvalItsmAudit = async () => {
  const { order_id, itsm_ticket_id } = data;
  const { state_id } = data.current_steps[0];
  const { approval, remark } = formModel;
  const operator = userStore.username;
  const params = { order_id, itsm_ticket_id, state_id, operator, approval, remark };

  await http.post(`/api/v1/woa/${getBusinessApiPath()}task/audit/apply/ticket`, params);
  getData(props.orderId, props.bkBizId);
  resetForm();
};

// 撤单操作
const isCancelItsmTicketLoading = ref(false);
const cancelItsmTicket = async () => {
  const { order_id } = data;
  isCancelItsmTicketLoading.value = true;
  try {
    await http.post(`/api/v1/woa/${getBusinessApiPath()}task/apply/ticket/itsm_audit/cancel`, { order_id });
    Message({ theme: 'success', message: t('撤单成功') });
    getData(props.orderId, props.bkBizId);
  } catch (error) {
    console.error(error);
  } finally {
    isCancelItsmTicketLoading.value = false;
  }
};

const refreshTask = useTimeoutPoll(
  () => {
    getData(props.orderId, props.bkBizId);
  },
  30000,
  { max: 60 },
);

onBeforeMount(() => {
  getData(props.orderId, props.bkBizId);
});
</script>

<template>
  <ticket-audit
    class="itsm-ticket-audit"
    :title="t('ITSM平台审批')"
    :loading="loading"
    :ticket-link="data.itsm_ticket_link"
    :logs="renderLogs"
  >
    <!-- 提单人可以在“管理员审批”和“leader审批”两个状态下进行撤单操作 -->
    <template
      #header-end
      v-if="userStore.username === creator && ['管理员审批', 'leader审批'].includes(data.current_steps[0]?.name)"
    >
      <bk-pop-confirm
        :title="t('撤销单据')"
        :content="t('撤销单据后，将取消本次的资源申请！')"
        trigger="click"
        placement="top-end"
        @confirm="cancelItsmTicket"
      >
        <bk-button class="cancel-btn" theme="primary" :loading="isCancelItsmTicketLoading">{{ t('撤单') }}</bk-button>
      </bk-pop-confirm>
    </template>
  </ticket-audit>

  <!-- 审批操作 -->
  <bk-dialog v-model:is-show="isShow" :title="t('审批')" @confirm="approvalItsmAudit">
    <bk-form form-type="vertical" :model="formModel">
      <bk-form-item :label="t('审批意见')" property="approval" required>
        <bk-radio v-model="formModel.approval" :label="true">{{ t('同意') }}</bk-radio>
        <bk-radio v-model="formModel.approval" :label="false">{{ t('拒绝') }}</bk-radio>
      </bk-form-item>
      <bk-form-item :label="t('审批说明')" property="remark">
        <bk-input type="textarea" :rows="4" :maxlength="200" v-model="formModel.remark" />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<style scoped lang="scss">
.itsm-ticket-audit {
  padding: 0 16px;

  .cancel-btn {
    margin-left: 24px;
    min-width: 88px;
  }

  :deep(.i-timeline-tag) {
    font-size: 14px;
    color: $font-deep-color;

    .approval-btn {
      margin-left: 24px;
    }
  }

  :deep(.i-timeline-content) {
    font-size: 12px;
  }
}
</style>
