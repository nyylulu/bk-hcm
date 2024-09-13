// 服务请求 （运营管理下）单据管理中 资源预测 tab
import { defineComponent, ref } from 'vue';
import Search from '@/components/resource-plan/applications/list/search';
import Table from '@/components/resource-plan/applications/list/table';
import cssModule from './index.module.scss';

import type { IListTicketsParam } from '@/typings/resourcePlan';

export default defineComponent({
  setup() {
    const tableRef = ref(null);

    const handleSearch = (searchModel: Partial<IListTicketsParam>) => {
      tableRef.value.searchTableData(searchModel);
    };

    return () => (
      <>
        <section class={cssModule.home}>
          <Search onSearch={handleSearch} isBiz={false}></Search>
          <Table ref={tableRef} isBiz={false}></Table>
        </section>
      </>
    );
  },
});
