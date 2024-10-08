// 资源管理 (业务)下 资源预测
import { defineComponent, ref } from 'vue';
import Table from '@/components/resource-plan/resource-manage/list/table';
import Search from '@/components/resource-plan/resource-manage/list/search';
import cssModule from './index.module.scss';
import { IListResourcesDemandsParam } from '@/typings/resourcePlan';
import useFormModel from '@/hooks/useFormModel';
import dayjs from 'dayjs';

export default defineComponent({
  setup() {
    const tableRef = ref(null);

    const handleSearch = (searchModel: Partial<IListResourcesDemandsParam>) => {
      tableRef.value?.searchTableData(searchModel);
    };

    const { formModel: expectTimeRange, setFormValues: setExpectTimeRange } = useFormModel({
      start: dayjs().startOf('month').format('YYYY-MM-DD'),
      end: dayjs().endOf('month').format('YYYY-MM-DD'),
    });

    return () => (
      <>
        <section class={cssModule.home}>
          <Search
            isBiz={true}
            onSearch={handleSearch}
            onExpectTimeRangeChange={(range) => setExpectTimeRange(range)}></Search>
          <Table isBiz={true} ref={tableRef} expectTimeRange={expectTimeRange}></Table>
        </section>
      </>
    );
  },
});
