import { defineComponent, computed } from 'vue';
import Panel from '@/components/panel';
import { Button } from 'bkui-vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';

export default defineComponent({
  setup() {
    const renderTableData = [
      {
        order_number: 'req123',
      },
    ];

    const { columns, settings } = useColumns('account');

    const tableColumns = computed(() => {
      return [
        ...columns,
        {
          label: '操作',
          width: 120,
          render: () => (
            <div>
              <Button text theme={'primary'}>
                撤销
              </Button>
              <Button text theme={'primary'}>
                重新申请
              </Button>
            </div>
          ),
        },
      ];
    });

    const { CommonTable } = useTable({
      searchOptions: {},
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

    return () => (
      <Panel>
        <CommonTable></CommonTable>
      </Panel>
    );
  },
});
