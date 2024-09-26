import { defineComponent, onBeforeMount } from 'vue';
import { useI18n } from 'vue-i18n';
import Panel from '@/components/panel';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useResourcePlanStore } from '@/store';
import { useTable } from '@/hooks/useResourcePlanTable';
import { useRoute } from 'vue-router';
import { IPageQuery } from '@/typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      default: true,
    },
  },

  setup(props) {
    const { t } = useI18n();
    const route = useRoute();
    const { getListChangeLogs, getListChangeLogsByOrg } = useResourcePlanStore();
    const { columns, settings } = useColumns('adjustmentEntry');
    const { getBizsId } = useWhereAmI();

    const getData = (page: IPageQuery) => {
      const params = {
        page,
        crp_demand_id: +route.query.crpDemandId,
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
