// 资源管理 (业务)下 资源预测
import { defineComponent, reactive, ref } from 'vue';
import Table from '@/components/resource-plan/resource-manage/list/table';
import Search from '@/components/resource-plan/resource-manage/list/search';
import cssModule from './index.module.scss';
import { IListResourcesDemandsParam } from '@/typings/resourcePlan';
import dayjs from 'dayjs';

export default defineComponent({
  setup() {
    const tableRef = ref(null);

    const handleSearch = (searchModel: Partial<IListResourcesDemandsParam>) => {
      tableRef.value?.searchTableData(searchModel);
    };

    const expectTimeRange = reactive({
      start: dayjs().startOf('month').subtract(1, 'week').startOf('day').format('YYYY-MM-DD'),
      end: dayjs().add(14, 'week').endOf('day').format('YYYY-MM-DD'),
    });

    return () => (
      <>
        <section class={cssModule.home}>
          <Search isBiz={true} v-model:expectTimeRange={expectTimeRange} onSearch={handleSearch}></Search>
          <Table isBiz={true} ref={tableRef} expectTimeRange={expectTimeRange}></Table>
        </section>
      </>
    );
  },
});
