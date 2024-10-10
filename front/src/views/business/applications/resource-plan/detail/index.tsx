import { defineComponent, ref, onBeforeMount } from 'vue';

import { useRoute } from 'vue-router';
import { useResourcePlanStore } from '@/store';

import Header from '@/components/resource-plan/applications/detail/header';
import Approval from '@/components/resource-plan/applications/detail/approval';
import Basic from '@/components/resource-plan/applications/detail/basic';
import List from '@/components/resource-plan/applications/detail/list';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { TicketByIdResult } from '@/typings/resourcePlan';

import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    const route = useRoute();
    const resourcePlanStore = useResourcePlanStore();
    const { getBizsId } = useWhereAmI();

    const ticketDetail = ref<TicketByIdResult>();
    const isLoading = ref(false);
    const errorMessage = ref();

    const getResultData = async () => {
      try {
        isLoading.value = true;
        const res = await resourcePlanStore.getBizResourcesTicketsById(getBizsId(), route.query?.id as string);

        if (res.data?.status_info?.status === 'failed') {
          errorMessage.value = res.data?.status_info?.message;
        } else {
          errorMessage.value = '';
        }
        ticketDetail.value = res?.data;
      } catch (error) {
        console.error('error', error); // eslint-disable-line no-console
      } finally {
        isLoading.value = false;
      }
    };

    onBeforeMount(getResultData);

    return () => (
      <bk-loading loading={isLoading.value}>
        <Header id={ticketDetail.value?.id} isBiz={true}></Header>
        <section class={cssModule.home}>
          <Approval
            statusInfo={ticketDetail.value?.status_info}
            class={cssModule['mb-16']}
            isBiz={true}
            errorMessage={errorMessage.value}></Approval>
          <Basic baseInfo={ticketDetail.value?.base_info} class={cssModule['mb-16']} isBiz={true}></Basic>
          <List demands={ticketDetail.value?.demands} isBiz={true}></List>
        </section>
      </bk-loading>
    );
  },
});
