import { defineComponent, ref, onBeforeMount } from 'vue';

import { useRoute } from 'vue-router';
import { useResourcePlanStore } from '@/store';

import Header from '@/components/resource-plan/applications/detail/header';
import Approval from '@/components/resource-plan/applications/detail/approval';
import Basic from '@/components/resource-plan/applications/detail/basic';
import List from '@/components/resource-plan/applications/detail/list';

import type { TicketByIdResult } from '@/typings/resourcePlan';

import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    const route = useRoute();
    const resourcePlanStore = useResourcePlanStore();

    const ticketDetail = ref<TicketByIdResult>();
    const isLoading = ref(false);

    const getResultData = async () => {
      try {
        isLoading.value = true;
        const res = await resourcePlanStore.getOpResourcesTicketsById(route.query?.id as string);
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
        <Header id={ticketDetail.value?.id} isBiz={false}></Header>
        <section class={cssModule.home}>
          <Approval statusInfo={ticketDetail.value?.status_info} class={cssModule['mb-16']} isBiz={false}></Approval>
          <Basic baseInfo={ticketDetail.value?.base_info} class={cssModule['mb-16']} isBiz={false}></Basic>
          <List demands={ticketDetail.value?.demands} isBiz={false}></List>
        </section>
      </bk-loading>
    );
  },
});
