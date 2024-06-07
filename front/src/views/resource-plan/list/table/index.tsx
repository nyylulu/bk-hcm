import { defineComponent, computed, PropType, watch } from 'vue';
import { Button } from 'bkui-vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
// import { useResourcePlanStore } from '@/store';

import { Plus as PlusIcon } from 'bkui-vue/lib/icon';
import cssModule from '../index.module.scss';
import Panel from '@/components/panel';
import { useRouter } from 'vue-router';

export default defineComponent({
  props: {
    searchData: {
      type: Object as PropType<{
        bk_biz_ids: number[];
        obs_projects: string[];
        ticket_ids: string;
        applicants: string[];
        submit_time_range: {
          start: string;
          end: string;
        };
      }>,
    },
  },
  setup(props) {
    const renderTableData = [
      {
        id: '999',
        forecast_order: '12123',
        demand_year_month: '2022-12-09',
        business: '业务名称',
        operation_product: '互娱运营支撑产品',
        creation_time: '2023-09-11',
      },
    ];

    // const resourcePlanStore = useResourcePlanStore();
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
          render: () => (
            <div>
              <Button text theme={'primary'}>
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
        query: { id: data.id },
      });
    };

    const getTableData = async () => {
      try {
        // const res = await resourcePlanStore.reqListTickets({
        //   ...props.searchData,
        // });
      } catch (error) {
        console.error(error, 'error'); // eslint-disable-line no-console
      }
    };

    watch(
      () => props.searchData,
      () => {
        getTableData();
      },
      {
        deep: true,
      },
    );
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
