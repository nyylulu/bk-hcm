// 资源管理 (业务)下 资源预测

import { defineComponent } from 'vue';
import Table from '@/components/resource-plan/resource-manage/list/table';
import Search from '@/components/resource-plan/resource-manage/list/search';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    return () => (
      <>
        <section class={cssModule.home}>
          <Search isBiz={true}></Search>
          <Table isBiz={true}></Table>
        </section>
      </>
    );
  },
});
