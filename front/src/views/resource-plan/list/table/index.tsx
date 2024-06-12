import { defineComponent, computed, onBeforeMount } from 'vue';

import { Button } from 'bkui-vue';
import { Plus as PlusIcon } from 'bkui-vue/lib/icon';

import { useRouter } from 'vue-router';
import { useTable } from '@/hooks/useResourcePlanTable';
import { useResourcePlanStore } from '@/store';
import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import Panel from '@/components/panel';

import cssModule from './index.module.scss';

import type { IListTicketsParam, IListTicketsResult } from '@/typings/resourcePlan';
import type { IPageQuery } from '@/typings';

export default defineComponent({
  setup(_, { expose }) {
    let searchModel: Partial<IListTicketsParam> = undefined;

    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();
    const { columns, settings } = useColumns('forecastDemand');
    const router = useRouter();

    const tableColumns = computed(() => {
      return [
        {
          label: t('预测单号'),
          field: 'forecast_order',
          isFormItem: true,
          render: ({ data }: { data: IListTicketsResult['detail'][0] }) => (
            <Button text theme='primary' onClick={() => handleToDetail(data)}>
              {data.id}
            </Button>
          ),
        },
        ...columns,
      ];
    });

    const getData = (page: IPageQuery) => {
      return resourcePlanStore.reqListTickets({
        page,
        ...searchModel,
      });
    };

    const {
      tableData,
      pagination,
      isLoading,
      handlePageChange,
      handlePageSizeChange,
      handleSort,
      triggerApi,
      resetPagination,
    } = useTable(getData);

    const handleToAdd = () => {
      router.push({ path: '/resource-plan/add' });
    };

    const handleToDetail = (data: IListTicketsResult['detail'][0]) => {
      router.push({
        path: '/resource-plan/detail',
        query: { id: data.id },
      });
    };

    const searchTableData = (data: Partial<IListTicketsParam>) => {
      searchModel = data;
      resetPagination();
      triggerApi();
    };

    onBeforeMount(triggerApi);

    expose({
      searchTableData,
    });

    return () => (
      <Panel>
        <Button theme='primary' onClick={handleToAdd} class={cssModule.button}>
          <PlusIcon class={cssModule['plus-icon']} />
          {t('新增')}
        </Button>
        <bk-loading loading={isLoading.value}>
          <bk-table
            remote-pagination
            show-overflow-tooltip
            data={tableData.value}
            pagination={pagination.value}
            columns={tableColumns.value}
            settings={settings.value}
            onPageLimitChange={handlePageSizeChange}
            onPageValueChange={handlePageChange}
            onColumnSort={handleSort}
          />
        </bk-loading>
      </Panel>
    );
  },
});
