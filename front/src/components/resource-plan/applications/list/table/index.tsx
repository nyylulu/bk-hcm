import { defineComponent, computed, onBeforeMount } from 'vue';

import { Button } from 'bkui-vue';

import { useRouter } from 'vue-router';
import { useTable } from '@/hooks/useResourcePlanTable';
import { useResourcePlanStore } from '@/store';
import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import Panel from '@/components/panel';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import type { IBizResourcesTicketsParam, IOpResourcesTicketsParam, IResourcesTicketItem } from '@/typings/resourcePlan';
import type { IPageQuery } from '@/typings';

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      required: true,
    },
  },
  setup(props, { expose }) {
    let searchModel: Partial<IBizResourcesTicketsParam | IOpResourcesTicketsParam>;

    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();
    const { columns, settings } = useColumns('receiptForecastDemand');
    const router = useRouter();
    const { getBizsId } = useWhereAmI();

    const tableColumns = computed(() => {
      const orderItem = {
        label: t('预测单号'),
        field: 'id',
        isFormItem: true,
        render: ({ data }: { data: IResourcesTicketItem }) => (
          <Button text theme='primary' onClick={() => handleToDetail(data)}>
            {data.id}
          </Button>
        ),
      };
      if (props.isBiz) {
        return [orderItem, ...columns];
      }
      return [
        orderItem,
        ...columns.slice(0, 2),
        {
          label: '业务',
          field: 'bk_biz_name',
          isDefaultShow: true,
        },
        {
          label: t('运营产品'),
          field: 'op_product_name',
          isDefaultShow: true,
        },
        {
          label: t('规划产品'),
          field: 'plan_product_name',
          isDefaultShow: true,
        },
        ...columns.slice(2),
      ];
    });

    const getData = (page: IPageQuery) => {
      if (props.isBiz) {
        return resourcePlanStore.getBizResourcesTicketsList(getBizsId(), {
          page,
          ...searchModel,
        });
      }

      return resourcePlanStore.getOpResourcesTicketsList({
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

    const handleToDetail = (data: IResourcesTicketItem) => {
      if (props.isBiz) {
        router.push({
          path: '/business/applications/resource-plan/detail',
          query: { id: data.id },
        });
      } else {
        router.push({
          path: '/service/my-apply/resource-plan/detail',
          query: { id: data.id },
        });
      }
    };

    const searchTableData = (data: Partial<IBizResourcesTicketsParam | IOpResourcesTicketsParam>) => {
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
