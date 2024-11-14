import { withDirectives, Ref } from 'vue';
import { Button, Dropdown, Message, bkTooltips } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { VendorEnum } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import Confirm, { confirmInstance } from '@/components/confirm';
import defaultUseColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import HostOperations, {
  OperationActions,
  operationMap,
} from '@/views/business/host/children/host-operations/internal/index';
import { OperationActions as DefaultOperationActions } from '@/views/business/host/children/host-operations/index';
import useSingleOperation from '@/views/business/host/children/host-operations/internal/use-single-operation';
import defaultUseSingleOperation from '@/views/business/host/children/host-operations/use-single-operation';
import defaultUseTableListQuery from '@/hooks/useTableListQuery';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useBusinessStore } from '@/store/business';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';
import type { PropsType } from '@/hooks/useTableListQuery';
import type { PluginHandlerType } from '../business-host-manage';

const { DropdownMenu, DropdownItem } = Dropdown;

type UseColumnsParams = {
  type?: string;
  isSimpleShow?: boolean;
  vendor?: string;
  extra?: {
    isLoading: Ref<boolean>;
    triggerApi: () => void;
    getHostOperationRef: () => any;
    getTableRef: () => any;
  };
};

const useColumns = ({ type = 'businessHostColumns', isSimpleShow = false, extra }: UseColumnsParams) => {
  const { t } = useI18n();
  const router = useRouter();
  const { getBizsId } = useWhereAmI();
  const businessStore = useBusinessStore();

  const { handleOperate: defaultHandleOperate } = defaultUseSingleOperation({
    beforeConfirm() {
      extra.isLoading.value = true;
    },
    confirmSuccess(type: string) {
      Message({ message: t('操作成功'), theme: 'success' });
      extra.triggerApi();
      if (type === OperationActions.RECYCLE) {
        router.push({ name: 'businessRecyclebin' });
      } else {
        extra.triggerApi();
      }
    },
    confirmComplete() {
      extra.isLoading.value = false;
    },
  });

  const { currentOperateRowIndex, getOperationConfig } = useSingleOperation({
    customOperate(type: OperationActions, data: any) {
      if (data.vendor === VendorEnum.ZIYAN) {
        if (type === OperationActions.RECYCLE) {
          // 使用批量回收操作
          extra.getTableRef()?.value?.clearSelection?.();
          extra.getHostOperationRef()?.value?.handleSingleZiyanRecycle?.(data);
        } else if (type === OperationActions.RESET) {
          // 重装单个
          extra.getHostOperationRef()?.value?.hostBatchResetDialogRef?.show([data.id]);
        } else {
          // 开机、关机、重启操作
          const { label } = operationMap[type];
          Confirm(`确定${label}`, <>当前操作主机为：{data.name}</>, async () => {
            confirmInstance.hide();
            extra.isLoading.value = true;
            try {
              const result = await businessStore.cvmOperateAsync(type, { ids: [data.id] });

              Message({ message: t('操作成功'), theme: 'success' });

              // 跳转至新任务详情页
              routerAction.redirect({
                name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
                params: { resourceType: ResourceTypeEnum.CVM, id: result.data.task_management_id },
                query: { bizs: getBizsId() },
              });
            } finally {
              extra.isLoading.value = false;
            }
          });
        }
      } else {
        defaultHandleOperate(type as DefaultOperationActions, data);
      }
    },
  });

  const operationDropdownList = Object.entries(operationMap)
    .filter(([type]) => ![OperationActions.RECYCLE, OperationActions.NONE].includes(type as OperationActions))
    .map(([type, value]) => ({
      type,
      label: value.label,
    }));

  const { columns, generateColumnsSettings } = defaultUseColumns(type, isSimpleShow);

  return {
    columns: [
      ...columns,
      {
        label: '操作',
        width: 120,
        showOverflowTooltip: false,
        render: ({ data, index }: { data: any; index: number }) => {
          return (
            <div class={'operation-column'}>
              {[
                withDirectives(
                  <Button
                    text
                    theme={'primary'}
                    class={`mr10 ${
                      getOperationConfig(OperationActions.RECYCLE, data).noPermission ? 'hcm-no-permision-text-btn' : ''
                    }`}
                    onClick={getOperationConfig(OperationActions.RECYCLE, data).clickHandler}
                    disabled={getOperationConfig(OperationActions.RECYCLE, data).disabled}>
                    {operationMap[OperationActions.RECYCLE].label}
                  </Button>,
                  [[bkTooltips, getOperationConfig(OperationActions.RECYCLE, data).tooltips]],
                ),
                <Dropdown
                  trigger='click'
                  popoverOptions={{
                    renderType: 'shown',
                    onAfterShow: () => (currentOperateRowIndex.value = index),
                    onAfterHidden: () => (currentOperateRowIndex.value = -1),
                  }}>
                  {{
                    default: () => (
                      <div
                        class={[`more-action${currentOperateRowIndex.value === index ? ' current-operate-row' : ''}`]}>
                        <i class={'hcm-icon bkhcm-icon-more-fill'}></i>
                      </div>
                    ),
                    content: () => (
                      <DropdownMenu>
                        {operationDropdownList.map(({ label, type }) => {
                          const { disabled, tooltips, clickHandler } = getOperationConfig(
                            type as OperationActions,
                            data,
                          );
                          return withDirectives(
                            <DropdownItem
                              key={type}
                              onClick={clickHandler}
                              extCls={`more-action-item${disabled ? ' disabled' : ''}`}>
                              {label}
                            </DropdownItem>,
                            [[bkTooltips, tooltips]],
                          );
                        })}
                      </DropdownMenu>
                    ),
                  }}
                </Dropdown>,
              ]}
            </div>
          );
        },
      },
    ],
    generateColumnsSettings,
  };
};

const useTableListQuery = (
  props: PropsType,
  type = 'cvms',
  completeCallback: () => void,
  apiMethod?: Function,
  apiName = 'list',
  args: any = {},
  extraResolveData?: (...args: any) => Promise<any>,
) => {
  return defaultUseTableListQuery(props, type, completeCallback, apiMethod, apiName, args, extraResolveData);
};

const pluginHandler: PluginHandlerType = {
  useColumns,
  useTableListQuery,
  HostOperations,
};

export default pluginHandler;
