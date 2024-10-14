// 资源管理 (业务)下 单据管理 tab 资源预测

import { defineComponent, ref } from 'vue';
import Search from '@/components/resource-plan/applications/list/search';
import Table from '@/components/resource-plan/applications/list/table';
import cssModule from './index.module.scss';

import type { IBizResourcesTicketsParam, IOpResourcesTicketsParam } from '@/typings/resourcePlan';

export default defineComponent({
  setup() {
    const tableRef = ref(null);

    const handleSearch = (searchModel: Partial<IBizResourcesTicketsParam | IOpResourcesTicketsParam>) => {
      tableRef.value.searchTableData(searchModel);
    };

    return () => (
      <section class={cssModule.home}>
        <Search onSearch={handleSearch} isBiz={true}></Search>
        <Table ref={tableRef} isBiz={true}></Table>
      </section>
    );
  },
});
