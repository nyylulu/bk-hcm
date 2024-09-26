// 资源管理 (业务)下 资源预测
import { defineComponent, ref } from 'vue';
import Table from '@/components/resource-plan/resource-manage/list/table';
import Search from '@/components/resource-plan/resource-manage/list/search';
import cssModule from './index.module.scss';
import { IListResourcesDemandsParam } from '@/typings/resourcePlan';

export default defineComponent({
  setup() {
    const tableRef = ref(null);

    const handleSearch = (searchModel: Partial<IListResourcesDemandsParam>) => {
      tableRef.value?.searchTableData(searchModel);
    };

    return () => (
      <>
        <section class={cssModule.home}>
          <Search isBiz={true} onSearch={handleSearch}></Search>
          <Table isBiz={true} ref={tableRef}></Table>
        </section>
      </>
    );
  },
});
