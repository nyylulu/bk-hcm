import { defineComponent, ref, onBeforeMount, computed } from 'vue';

import { useRoute } from 'vue-router';
import { useResourcePlanStore } from '@/store';

import Header from '@/components/resource-plan/applications/detail/header';
import Approval from '@/components/resource-plan/applications/detail/approval';
import TicketAudit from '@/components/resource-plan/applications/detail/ticket-audit/index.vue';
import Basic from '@/components/resource-plan/applications/detail/basic';
import List from '@/components/resource-plan/applications/detail/list';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { IPlanTicketAudit, TicketByIdResult } from '@/typings/resourcePlan';

import cssModule from './index.module.scss';
import useTimeoutPoll from '@/hooks/use-timeout-poll';

export default defineComponent({
  setup() {
    const route = useRoute();
    const resourcePlanStore = useResourcePlanStore();
    const { getBizsId } = useWhereAmI();

    const ticketDetail = ref<TicketByIdResult>();
    const ticketAuditDetail = ref<IPlanTicketAudit>();
    const isLoading = ref(false);
    const errorMessage = ref();

    const isTicketAuditDetailShow = computed(() => ticketAuditDetail.value?.itsm_audit.status !== 'init');

    const getResultData = async () => {
      try {
        isLoading.value = true;
        const promise = Promise.all([
          resourcePlanStore.getBizResourcesTicketsById(getBizsId(), route.query?.id as string),
          resourcePlanStore.getBizResourcesTicketsAuditById(getBizsId(), route.query?.id as string),
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
        <Header id={ticketDetail.value?.id} isBiz={true}></Header>
        <section class={cssModule.home}>
          <Approval
            statusInfo={ticketDetail.value?.status_info}
            class={cssModule['mb-16']}
            isBiz={true}
            errorMessage={errorMessage.value}
            ticketAuditDetail={ticketAuditDetail.value}></Approval>
          {isTicketAuditDetailShow.value && <TicketAudit detail={ticketAuditDetail.value} />}
          <Basic baseInfo={ticketDetail.value?.base_info} class={cssModule['mb-16']} isBiz={true}></Basic>
          <List demands={ticketDetail.value?.demands} isBiz={true}></List>
        </section>
      </bk-loading>
    );
  },
});
