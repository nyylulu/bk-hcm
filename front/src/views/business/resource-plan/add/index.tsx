import { defineComponent, ref, provide, watch, nextTick } from 'vue';
import { useRoute } from 'vue-router';
import planRemark from './plan-remark.js';
import cssModule from './index.module.scss';
import Header from './header';
import Basic from './basic';
import List from './list';
import Memo from './memo';
import Button from './button';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { IPlanTicket, IPlanTicketDemand } from '@/typings/resourcePlan';
import Add from '@/components/resource-plan/add';

export default defineComponent({
  setup() {
    const route = useRoute();
    const { getBizsId } = useWhereAmI();

    const basicRef = ref();
    const listRef = ref();
    const memoRef = ref();
    const isShowAdd = ref(false);
    const initDemand = ref();
    const planTicket = ref<IPlanTicket>({
      bk_biz_id: getBizsId(),
      demand_class: 'CVM',
      remark: planRemark,
      demands: [],
    });
    const initAddParams = ref({});

    const handleShowAdd = () => {
      initDemand.value = undefined;
      isShowAdd.value = true;
    };

    const handleShowModify = (data: IPlanTicketDemand) => {
      initDemand.value = data;
      isShowAdd.value = true;
    };

    const validate = () => {
      return Promise.all([basicRef.value.validate(), listRef.value.validate(), memoRef.value.validate()]);
    };

    watch(
      () => route.query.action,
      (action) => {
        if (action === 'add') {
          initAddParams.value = JSON.parse(decodeURIComponent(route.query.payload as string));
          nextTick(() => {
            handleShowAdd();
          });
        }
      },
      { immediate: true },
    );

    provide('validate', validate);

    return () => (
      <>
        <Header></Header>
        <section class={cssModule.home}>
          <Basic v-model={planTicket.value} ref={basicRef}></Basic>
          <List
            class={cssModule['mt-16']}
            ref={listRef}
            v-model={planTicket.value}
            onShow-add={handleShowAdd}
            onShow-modify={handleShowModify}></List>
          <Memo class={cssModule['mt-16']} ref={memoRef} v-model={planTicket.value}></Memo>
          <Button class={cssModule['mt-16']} v-model={planTicket.value}></Button>
        </section>
        <Add
          v-model:isShow={isShowAdd.value}
          v-model={planTicket.value}
          initDemand={initDemand.value}
          initAddParams={initAddParams.value}></Add>
      </>
    );
  },
});
