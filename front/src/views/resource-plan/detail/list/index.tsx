import { defineComponent, PropType } from 'vue';
import Panel from '@/components/panel';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
// import { Button } from 'bkui-vue';
import cssModule from '../index.module.scss';
import { TicketDemands } from '@/typings/resourcePlan';
export default defineComponent({
  props: {
    tableData: {
      type: Array as PropType<TicketDemands[]>,
    },
  },
  setup(props) {
    const { columns, settings } = useColumns('forecastDemandDetail');

    const { CommonTable } = useTable({
      searchOptions: {
        disabled: true,
      },
      tableOptions: {
        columns,
        reviewData: props?.tableData || [],
        extra: {
          settings: settings.value,
        },
      },
      requestOption: {
        type: '',
      },
    });
    return () => (
      <Panel title='资源预测' class={cssModule.relative}>
        {/* <Button theme='primary' outline class={cssModule['pre-btn']}>
          调整需求
        </Button> */}
        <CommonTable></CommonTable>
      </Panel>
    );
  },
});
