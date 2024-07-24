import { withDirectives, Ref } from 'vue';
import { Button, Dropdown, Message, bkTooltips } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import { VendorEnum } from '@/common/constant';
import defaultUseColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import HostOperations, {
  OperationActions,
  operationMap,
} from '@/views/business/host/children/host-operations/internal/index';
import useSingleOperation from '@/views/business/host/children/host-operations/internal/use-single-operation';
import defaultUseSingleOperation from '@/views/business/host/children/host-operations/use-single-operation';
import defaultUseTableListQuery from '@/hooks/useTableListQuery';
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
  const { isOperateDisabled, getMenuTooltipsOption, currentOperateRowIndex } = useSingleOperation();
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

  const handleOperate = async (type: OperationActions, data: any) => {
    if (isOperateDisabled(type, data)) return;

    if (data.vendor === VendorEnum.ZIYAN) {
      // 使用批量回收操作
      extra.getTableRef()?.value?.clearSelection?.();
      extra.getHostOperationRef()?.value?.handleSingleZiyanRecycle?.(data);
    } else {
      defaultHandleOperate(type, data);
    }
  };

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
                    class={'mr10'}
                    onClick={() => handleOperate(OperationActions.RECYCLE, data)}
                    disabled={isOperateDisabled(OperationActions.RECYCLE, data)}>
                    {operationMap[OperationActions.RECYCLE].label}
                  </Button>,
                  [[bkTooltips, getMenuTooltipsOption(OperationActions.RECYCLE, data)]],
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
                          return withDirectives(
                            <DropdownItem
                              key={type}
                              onClick={() => handleOperate(type as OperationActions, data)}
                              extCls={`more-action-item${
                                isOperateDisabled(type as OperationActions, data) ? ' disabled' : ''
                              }`}>
                              {label}
                            </DropdownItem>,
                            [[bkTooltips, getMenuTooltipsOption(type as OperationActions, data)]],
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
  apiMethod?: Function,
  apiName = 'list',
  args: any = {},
  extraResolveData?: (...args: any) => Promise<any>,
) => {
  return defaultUseTableListQuery(props, type, apiMethod, apiName, args, extraResolveData);
};

const pluginHandler: PluginHandlerType = {
  useColumns,
  useTableListQuery,
  HostOperations,
};

export default pluginHandler;
