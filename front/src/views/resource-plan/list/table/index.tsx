import { defineComponent, computed } from 'vue';
import { Button } from 'bkui-vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';

import { Plus as PlusIcon } from 'bkui-vue/lib/icon';
import cssModule from '../index.module.scss';
import Panel from '@/components/panel';
import { useRouter } from 'vue-router';

export default defineComponent({
  setup() {
    const renderTableData = [
      {
        forecast_order: '12123',
        demand_year_month: '2022-12-09',
        business: '业务名称',
        operation_product: '互娱运营支撑产品',
        creation_time: '2023-09-11',
      },
    ];
    const { columns, settings } = useColumns('forecastDemand');
    const router = useRouter();

    const tableColumns = computed(() => {
      return [
        {
          label: '预测单号',
          field: 'forecast_order',
          isFormItem: true,
          render: ({ data }) => (
            <Button text theme='primary' onClick={() => handleToDetail(data)}>
              {data.forecast_order}
            </Button>
          ),
        },
        ...columns,
        {
          label: '操作',
          width: 120,
          render: ({ data }: any) => (
            <div>
              <Button text theme={'primary'} onClick={() => handleAdjustmentDemand(data)}>
                调整需求
              </Button>
            </div>
          ),
        },
      ];
    });

    const { CommonTable } = useTable({
      searchOptions: {
        disabled: true,
      },
      tableOptions: {
        columns: tableColumns.value,
        reviewData: renderTableData,
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: '',
      },
    });

    // const formItemBase = tableColumns.value.filter((item) => item.isFormItem);

    const handleToAdd = () => {
      router.push({ path: '/resource-plan/add' });
    };

    const handleToDetail = (data) => {
      router.push({
        path: '/resource-plan/detail',
        query: { data: JSON.stringify(data) },
      });
      // baseMap.value = formItemBase.reduce((pre, cur) => {
      //   if (cur.field in data) {
      //     pre.push({
      //       label: cur.label,
      //       value: data[cur.field],
      //     });
      //   }
      //   return pre;
      // }, []);
    };

    const handleAdjustmentDemand = () => {};
    return () => (
      <Panel>
        <Button theme='primary' onClick={() => handleToAdd()}>
          <PlusIcon class={cssModule['add-font']} />
          新增
        </Button>
        {/* 资源预测表格 */}
        <CommonTable></CommonTable>
      </Panel>
    );
  },
});
