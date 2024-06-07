import { defineComponent, ref, onBeforeMount } from 'vue';
import Approval from './approval';
import Basic from './basic';
import List from './list';
import cssModule from './index.module.scss';
import { useResourcePlanStore } from '@/store';
import { useRoute } from 'vue-router';

export default defineComponent({
  setup() {
    const route = useRoute();
    const resourcePlanStore = useResourcePlanStore();

    const baseData = ref();
    const tableData = ref(undefined);
    const getResultData = async () => {
      try {
        const res = await resourcePlanStore.getTicketById(route.query?.id as string);
        const { base_info: baseInfo, demands } = res?.data;
        baseData.value = baseInfo;
        tableData.value = demands;
      } catch (error) {
        console.error('error', error); // eslint-disable-line no-console
      }
    };
    onBeforeMount(getResultData);
    return () => (
      <section class={cssModule.home}>
        <Approval></Approval>
        <Basic baseData={baseData.value}></Basic>
        <List tableData={tableData.value}></List>
      </section>
    );
  },
});
