import { defineComponent, ref } from 'vue';
import Search from './search';
import Table from './table';
import cssModule from './index.module.scss';

import type { IListTicketsParam } from '@/typings/resourcePlan';

export default defineComponent({
  setup() {
    const tableRef = ref(null);

    const handleSearch = (searchModel: Partial<IListTicketsParam>) => {
      tableRef.value.searchTableData(searchModel);
    };

    return () => (
      <section class={cssModule.home}>
        <Search onSearch={handleSearch}></Search>
        <Table ref={tableRef}></Table>
      </section>
    );
  },
});
