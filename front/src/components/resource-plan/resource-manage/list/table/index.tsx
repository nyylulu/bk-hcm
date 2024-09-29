import { defineComponent, computed, onBeforeMount, ref } from 'vue';
import { Button, Dropdown } from 'bkui-vue';
import { Plus as PlusIcon } from 'bkui-vue/lib/icon';
import { useTable } from '@/hooks/useResourcePlanTable';
import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import Panel from '@/components/panel';
import BatchCancellationDialog from '@/components/resource-plan/resource-manage/list/table/components/batch-cancellation-dialog/batch-cancellation-dialog';
import cssModule from './index.module.scss';
import { useRoute, useRouter } from 'vue-router';
import { IListResourcesDemandsItem, IListResourcesDemandsParam, ResourcesDemandsStatus } from '@/typings/resourcePlan';
import { IPageQuery } from '@/typings';
import { useResourcePlanStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';

const { DropdownMenu, DropdownItem } = Dropdown;

export enum OperationActions {
  EDIT = 'edit',
  DELETE = 'delete',
  ADJUST = 'adjust',
  CANCEL = 'cancel',
  PURCHASE = 'purchase',
}

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      default: true,
    },
  },
  setup(props, { expose }) {
    let searchModel: Partial<IListResourcesDemandsParam> = undefined;

    const { t } = useI18n();
    const route = useRoute();
    const router = useRouter();
    const { columns, generateColumnsSettings } = useColumns('resourceForecast');
    const { getResourcesDemandsList, getResourcesDemandsListByOrg } = useResourcePlanStore();
    const { getBizsId } = useWhereAmI();

    const { authVerifyData, handleAuth } = useVerify();
    const globalPermissionDialog = useGlobalPermissionDialog();
    const operationMap = {
      [OperationActions.EDIT]: {
        label: t('修改'),
        loading: false,
      },
      [OperationActions.DELETE]: {
        label: t('删除'),
        loading: false,
      },
      [OperationActions.ADJUST]: {
        label: t('调整'),
        loading: false,
      },
      [OperationActions.CANCEL]: {
        label: t('取消'),
        loading: false,
      },
    };

    const bizActions = [OperationActions.ADJUST, OperationActions.CANCEL];
    const serviceActions = [OperationActions.EDIT, OperationActions.DELETE];
    const operationDropdownList = Object.entries(operationMap)
      .filter(([type]) => (props.isBiz ? bizActions : serviceActions).includes(type as OperationActions))
      .map(([type, value]) => ({
        type,
        label: value.label,
      }));

    const tableRef = ref();
    const isShow = ref(false);
    const currentRowsData = ref<IListResourcesDemandsItem[]>([]);

    const selection = computed<IListResourcesDemandsItem[]>(() => tableRef.value?.getSelection?.() || []);
    const tableColumns = computed(() => {
      const newColumns = props.isBiz ? columns.slice(2) : columns.slice(0, -1);
      return [
        {
          type: 'selection',
          width: 30,
          minWidth: 30,
          fixed: 'left',
          onlyShowOnList: true,
        },
        {
          label: t('预测ID'),
          field: 'crp_demand_id',
          minWidth: 120,
          fixed: 'left',
          isDefaultShow: true,
          render: ({ data }: any) => {
            return (
              <Button
                theme='primary'
                text
                onClick={() => {
                  router.push({
                    path: props.isBiz ? '/business/resource-plan/detail' : '/service/resource-plan/detail',
                    query: { ...route.query, crpDemandId: data.crp_demand_id },
                  });
                }}>
                {data.crp_demand_id}
              </Button>
            );
          },
        },
        ...newColumns,
        {
          label: t('操作'),
          field: 'actions',
          fixed: 'right',
          minWidth: 100,
          isDefaultShow: true,
          render: ({ data }: { data: IListResourcesDemandsItem }) => {
            const firstActions = operationDropdownList.slice(0, 1)[0];
            return (
              <div class={cssModule['operation-column']}>
                <Button
                  text
                  theme={'primary'}
                  class={`${
                    !authVerifyData.value?.permissionAction?.biz_resource_plan_operate
                      ? 'hcm-no-permision-text-btn'
                      : undefined
                  }`}
                  onClick={() => handleOperate(firstActions.type as OperationActions, data)}>
                  {firstActions.label}
                </Button>
                <Dropdown trigger='click'>
                  {{
                    default: () => (
                      <div class={cssModule['more-action']}>
                        <i class={'hcm-icon bkhcm-icon-more-fill'}></i>
                      </div>
                    ),
                    content: () => (
                      <DropdownMenu>
                        {operationDropdownList.slice(1).map(({ label, type }) => (
                          <DropdownItem
                            key={type}
                            onClick={() => handleOperate(type as OperationActions, data)}
                            class={`${
                              !authVerifyData.value?.permissionAction?.biz_resource_plan_operate
                                ? 'hcm-no-permision-text-btn'
                                : undefined
                            }`}>
                            {label}
                          </DropdownItem>
                        ))}
                      </DropdownMenu>
                    ),
                  }}
                </Dropdown>
              </div>
            );
          },
        },
      ];
    });

    const settings = generateColumnsSettings(tableColumns.value);

    const getData = (page: IPageQuery) => {
      const params = {
        page,
        ...searchModel,
      };
      try {
        return props.isBiz ? getResourcesDemandsList(getBizsId(), params) : getResourcesDemandsListByOrg(params);
      } catch (error) {
        console.error('Error fetching data:', error);
      }
    };

    const {
      tableData,
      overview,
      pagination,
      isLoading,
      handlePageChange,
      handlePageSizeChange,
      handleSort,
      triggerApi,
      resetPagination,
    } = useTable(getData);

    const isRowSelectEnable = ({ row }: { row: IListResourcesDemandsItem }) => {
      return row.status === ResourcesDemandsStatus.CAN_APPLY;
    };

    const handleToAdd = () => {
      // 无权限
      if (!authVerifyData.value.permissionAction.biz_resource_plan_operate) {
        handleAuth('biz_resource_plan_operate');
        globalPermissionDialog.setShow(true);
      } else {
        router.push({
          path: '/business/resource-plan/add',
        });
      }
    };

    const handleToEdit = (data: IListResourcesDemandsItem[]) => {
      if (!authVerifyData.value.permissionAction.biz_resource_plan_operate) {
        // 无权限
        handleAuth('biz_resource_plan_operate');
        globalPermissionDialog.setShow(true);
      } else {
        const planIds = data.map(({ crp_demand_id }) => crp_demand_id).join(',');
        const path = props.isBiz ? '/business/service/resource-plan-mod' : '/service/resource-plan/mod';
        router.push({
          path,
          query: {
            planIds,
          },
        });
      }
    };

    const handleCancel = () => {
      if (!authVerifyData.value.permissionAction.resource_plan_update) {
        handleAuth('resource_plan_update');
        globalPermissionDialog.setShow(true);
      } else {
        currentRowsData.value = selection.value;
        isShow.value = true;
      }
    };

    const handleOperate = (type: OperationActions, data: IListResourcesDemandsItem) => {
      if (!authVerifyData.value?.permissionAction?.biz_resource_plan_operate) {
        // 无权限
        handleAuth('biz_resource_plan_operate');
        globalPermissionDialog.setShow(true);
        return;
      }
      if (type === OperationActions.CANCEL) {
        currentRowsData.value = [data];
        isShow.value = true;
      } else if (type === OperationActions.ADJUST) {
        handleToEdit([data]);
      }
    };

    const searchTableData = (data: Partial<IListResourcesDemandsParam>) => {
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
        <div class={cssModule['table-header']}>
          <div class={cssModule['toolbar-buttons']}>
            {props.isBiz && (
              <>
                <Button
                  class={`${cssModule.button} ${
                    !authVerifyData.value.permissionAction.biz_resource_plan_operate
                      ? 'hcm-no-permision-btn'
                      : undefined
                  }`}
                  theme='primary'
                  onClick={handleToAdd}>
                  <PlusIcon class={cssModule['plus-icon']} />
                  {t('新增预测')}
                </Button>
                <Button
                  class={`${cssModule.button} ${
                    !authVerifyData.value.permissionAction.biz_resource_plan_operate
                      ? 'hcm-no-permision-btn'
                      : undefined
                  }`}
                  onClick={() => handleToEdit(selection.value)}
                  disabled={!selection.value.length}>
                  {t('批量调整')}
                </Button>
                <Button
                  class={`${cssModule.button} ${
                    !authVerifyData.value.permissionAction.biz_resource_plan_operate
                      ? 'hcm-no-permision-btn'
                      : undefined
                  }`}
                  onClick={handleCancel}
                  disabled={!selection.value.length}>
                  {t('批量取消')}
                </Button>
              </>
            )}
          </div>
          <div class={cssModule.overview}>
            <div>
              <span>{`${t('本月即将过期 CPU ')}${overview.value?.expiring_cpu_core || '--'}${t('核')}`}</span>
            </div>
            <div class={cssModule['cpu-stats']}>
              <span>
                {t('CPU 总核数')}：
                <span class={cssModule.num}>
                  {overview.value?.total_cpu_core || '--'}/{overview.value?.total_applied_core || '--'}
                </span>
              </span>
              <span>
                {t('预测内')}：
                <span class={cssModule.num}>
                  {overview.value?.in_plan_cpu_core || '--'}/{overview.value?.in_plan_applied_cpu_core || '--'}
                </span>
              </span>
              <span>
                {t('预测外')}：
                <span class={cssModule.num}>
                  {overview.value?.out_plan_cpu_core || '--'}/{overview.value?.out_plan_applied_cpu_core || '--'}
                </span>
              </span>
            </div>
          </div>
        </div>
        <bk-loading loading={isLoading.value}>
          <bk-table
            ref={tableRef}
            row-hover='auto'
            remote-pagination
            show-overflow-tooltip
            data={tableData.value}
            pagination={pagination.value}
            columns={tableColumns.value}
            settings={settings.value}
            isRowSelectEnable={isRowSelectEnable}
            onPageLimitChange={handlePageSizeChange}
            onPageValueChange={handlePageChange}
            onColumnSort={handleSort}
          />
        </bk-loading>
        <BatchCancellationDialog v-model:isShow={isShow.value} data={currentRowsData.value} onRefresh={triggerApi} />
      </Panel>
    );
  },
});
