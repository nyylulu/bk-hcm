// 运营(服务)模块-单据管理-资源预测-详情
import { defineComponent, ref, onBeforeMount, computed } from 'vue';

import { useRoute } from 'vue-router';
import { useResourcePlanStore } from '@/store';

import Header from '@/components/resource-plan/applications/detail/header';
import Approval from '@/components/resource-plan/applications/detail/approval';
import ResourcePlanTicketAudit from '@/components/resource-plan/applications/detail/ticket-audit/index.vue';
import Basic from '@/components/resource-plan/applications/detail/basic';
import List from '@/components/resource-plan/applications/detail/list/index.vue';

import type { IPlanTicketAudit, TicketByIdResult } from '@/typings/resourcePlan';

import cssModule from './index.module.scss';
import useTimeoutPoll from '@/hooks/use-timeout-poll';

export default defineComponent({
  setup() {
    const route = useRoute();
    const resourcePlanStore = useResourcePlanStore();

    const ticketDetail = ref<TicketByIdResult>();
    const ticketAuditDetail = ref<IPlanTicketAudit>();
    const isLoading = ref(false);
    const errorMessage = ref();

    const isTicketAuditDetailShow = computed(() => ticketAuditDetail.value?.itsm_audit.status !== 'init');

    const getResultData = async () => {
      try {
        isLoading.value = true;
        const promise = Promise.all([
          resourcePlanStore.getOpResourcesTicketsById(route.query?.id as string),
          resourcePlanStore.getOpResourcesTicketsAuditById(route.query?.id as string),
        ]);
        const [res1, res2] = await promise;

        if (res1.data?.status_info?.status === 'failed') {
          errorMessage.value = res1.data?.status_info?.message;
        } else {
          errorMessage.value = '';
        }
        ticketDetail.value = res1?.data;
        ticketAuditDetail.value = res2?.data;

        // 如果处于 init、auditing 状态，每 30s 刷新一次
        if (res1?.data?.status_info?.status === 'init' || res1?.data?.status_info?.status === 'auditing') {
          autoFlashTask.resume();
        } else {
          autoFlashTask.pause();
        }
      } catch (error) {
        console.error('error', error); // eslint-disable-line no-console
      } finally {
        isLoading.value = false;
      }
    };

    const autoFlashTask = useTimeoutPoll(() => {
      getResultData();
    }, 30000);

    onBeforeMount(getResultData);

    return () => (
      <bk-loading loading={isLoading.value}>
        <Header id={ticketDetail.value?.id} isBiz={false}></Header>
        <section class={cssModule.home}>
          <Approval
            statusInfo={ticketDetail.value?.status_info}
            class={cssModule['mb-16']}
            isBiz={false}
            errorMessage={errorMessage.value}
            ticketAuditDetail={ticketAuditDetail.value}></Approval>
          {isTicketAuditDetailShow.value && (
            <ResourcePlanTicketAudit
              detail={ticketAuditDetail.value}
              fetchData={getResultData}
              timeoutPollAction={autoFlashTask}
            />
          )}
          <Basic baseInfo={ticketDetail.value?.base_info} class={cssModule['mb-16']} isBiz={false}></Basic>
          <List demands={ticketDetail.value?.demands} isBiz={false}></List>
        </section>
      </bk-loading>
    );
  },
});
