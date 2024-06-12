import { defineComponent, computed } from 'vue';
import Panel from '@/components/panel';
import { Button } from 'bkui-vue';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  setup() {
    const renderTableData = [
      {
        order_number: 'req123',
      },
    ];

    const { t } = useI18n();
    const { columns, settings } = useColumns('account');

    const tableColumns = computed(() => {
      return [
        ...columns,
        {
          label: t('操作'),
          width: 120,
          render: () => (
            <div>
              <Button text theme={'primary'}>
                {t('撤销')}
              </Button>
              <Button text theme={'primary'}>
                {t('重新申请')}
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
