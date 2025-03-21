import { defineComponent, onBeforeMount } from 'vue';
import { useI18n } from 'vue-i18n';
import Panel from '@/components/panel';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useResourcePlanStore } from '@/store';
import { useTable } from '@/hooks/useResourcePlanTable';
import { useRoute } from 'vue-router';
import { IPageQuery } from '@/typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

import { Button } from 'bkui-vue';
import { Column } from 'bkui-vue/lib/table/props';

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      default: true,
    },
    currentBusinessId: Number,
  },

  setup(props) {
    const { t } = useI18n();
    const route = useRoute();
    const { getListChangeLogs, getListChangeLogsByOrg } = useResourcePlanStore();
    const { columns, settings } = useColumns('adjustmentEntry');
    const { getBizsId } = useWhereAmI();

    // 预测单号列的render
    columns.find((item: Column) => item.field === 'ticket_id').render = ({ cell }: { cell: string }) => (
      <Button
        theme='primary'
        text
        onClick={() =>
          routerAction.redirect({
            name: 'BizInvoiceResourceDetail',
            query: { id: cell, [GLOBAL_BIZS_KEY]: props.currentBusinessId },
          })
        }>
        {cell}
      </Button>
    );

    const getData = (page: IPageQuery) => {
      const params = {
        page,
        demand_id: route.query.demandId as string,
      };
      try {
        return props.isBiz ? getListChangeLogs(getBizsId(), params) : getListChangeLogsByOrg(params);
      } catch (error) {
        console.error('Error fetching data:', error);
      }
    };

    const { tableData, pagination, isLoading, handlePageChange, handlePageSizeChange, triggerApi, resetPagination } =
      useTable(getData);

    onBeforeMount(() => {
      resetPagination();
      triggerApi();
    });

    return () => (
      <Panel title={t('调整记录')}>
        <bk-loading loading={isLoading.value}>
          <bk-table
            row-hover='auto'
            show-overflow-tooltip
            settings={settings.value}
            columns={columns}
            data={tableData.value}
            pagination={pagination.value}
            onPageLimitChange={handlePageSizeChange}
            onPageValueChange={handlePageChange}
          />
        </bk-loading>
      </Panel>
    );
  },
});
