import { defineComponent, ref, onBeforeMount } from 'vue';

import { useRoute } from 'vue-router';
import { useResourcePlanStore } from '@/store';

import Header from './header';
import Approval from './approval';
import Basic from './basic';
import List from './list';

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
        const res = await resourcePlanStore.getTicketById(route.query?.id as string);
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
        <Header id={ticketDetail.value?.id}></Header>
        <section class={cssModule.home}>
          <Approval statusInfo={ticketDetail.value?.status_info} class={cssModule['mb-16']}></Approval>
          <Basic baseInfo={ticketDetail.value?.base_info} class={cssModule['mb-16']}></Basic>
          <List demands={ticketDetail.value?.demands}></List>
        </section>
      </bk-loading>
    );
  },
});
