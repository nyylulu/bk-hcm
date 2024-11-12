import { PropType, defineComponent, ref, toRefs, withDirectives } from 'vue';
import { useRouter } from 'vue-router';
import { Button, Dialog, Dropdown, Loading, Message, bkTooltips } from 'bkui-vue';
import cssModule from '../index.module.scss';
import { AngleDown } from 'bkui-vue/lib/icon';
import { BkDropdownItem, BkDropdownMenu } from 'bkui-vue/lib/dropdown';
import { useZiyanScrStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import CommonLocalTable from '@/components/LocalTable';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import useBatchOperation from './use-batch-operation';
import { operationMap as defaultOperationMap, OperationMapItem } from '../index';
import RecycleFlow from './recycle-flow.vue';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';

export enum OperationActions {
  NONE = 'none',
  START = 'start',
  STOP = 'stop',
  REBOOT = 'reboot',
  RECYCLE = 'recycle',
  RESET = 'reset',
}

export const operationMap: Record<OperationActions, OperationMapItem> = {
  ...defaultOperationMap,
  [OperationActions.RESET]: {
    label: '重装',
    disabledStatus: [] as string[],
    loading: false,
    // 鉴权参数
    authId: 'biz_iaas_resource_delete',
    actionName: 'biz_iaas_resource_delete',
  },
};

export default defineComponent({
  props: {
    selections: {
      type: Array as PropType<
        Array<{
          status: string;
          id: string;
          vendor: string;
          private_ipv4_addresses: string[];
          __formSingleOp: boolean;
        }>
      >,
    },
    onFinished: {
      type: Function as PropType<(type: 'confirm' | 'cancel') => void>,
    },
  },
  setup(props, { expose }) {
    const dialogRef = ref(null);
    const recycleFlowRef = ref(null);

    const router = useRouter();
    const scrStore = useZiyanScrStore();
    const { getBizsId } = useWhereAmI();
    const { selections } = toRefs(props);

    const { authVerifyData, handleAuth } = useVerify();
    const globalPermissionDialog = useGlobalPermissionDialog();

    const {
      operationType,
      isDialogShow,
      isConfirmDisabled,
      operationsDisabled,
      computedTitle,
      computedTips,
      computedContent,
      isLoading,
      selected,
      isDialogLoading,
      computedColumns,
      tableData,
      searchData,
      isMix,
      isZiyanOnly,
      isZiyanRecycle,
      hostPrivateIP4s,
      selectedRowPrivateIPs,
      selectedRowPublicIPs,
      handleSwitch,
      handleConfirm,
      handleCancelDialog,
    } = useBatchOperation({
      selections,
      onFinished: props.onFinished,
    });

    const getOperationConfig = (type: OperationActions) => {
      // 点击事件（值缺省时，为默认点击事件）
      const clickHandler = () => handleClickMenu(type);

      if (isMix.value) {
        return {
          disabled: true,
          tooltips: { content: '腾讯云自研云和公有云的主机，不支持同时操作', disabled: false },
          clickHandler,
        };
      }

      // 非自研云不支持重装操作
      const isReset = type === OperationActions.RESET;
      if (!isZiyanOnly.value && isReset) {
        return {
          disabled: true,
          tooltips: { content: '暂不支持', disabled: false },
          clickHandler,
        };
      }

      // 预鉴权
      const { authId, actionName } = operationMap[type];
      const noPermission = !authVerifyData?.value?.permissionAction?.[authId];
      if (authId && actionName && noPermission) {
        return {
          disabled: false,
          tooltips: { disabled: true },
          clickHandler: () => {
            handleAuth(actionName);
            globalPermissionDialog.setShow(true);
          },
        };
      }

      return { disabled: false, tooltips: { disabled: true }, clickHandler };
    };

    const handleClickMenu = (type: OperationActions) => {
      if (getOperationConfig(type).disabled) {
        return;
      }
      operationType.value = type;
      // 主机重装操作
      if (type === OperationActions.RESET) {
        console.error('🚀 ~ handleClickMenu ~ type:', type);
      }
    };

    const ziyanRecycleSelected = ref([]);
    const handleZiyanRecycleSelectChange = (selected: any[]) => {
      ziyanRecycleSelected.value = selected;
    };

    const handleZiyanRecycleSubmit = async () => {
      try {
        isLoading.value = true;
        Message({ message: `${computedTitle.value}中, 请不要操作`, theme: 'warning', delay: 500 });
        if (recycleFlowRef.value?.isSelectionRecycleTypeChange) {
          const suborder_id_types = ziyanRecycleSelected.value.map((item) => ({
            suborder_id: item.suborder_id,
            recycle_type: item.recycle_type,
          }));
          await scrStore.startRecycleOrderByRecycleType({ suborder_id_types });
        } else {
          const orderIds = ziyanRecycleSelected.value.map((item) => item.order_id);
          await scrStore.startRecycleOrder({ order_id: orderIds });
        }
        Message({ message: '操作成功', theme: 'success' });
        props.onFinished?.('confirm');
        router.push({ name: 'ApplicationsManage', query: { bizs: getBizsId(), type: 'host_recycle' } });
        operationType.value = OperationActions.NONE;
      } catch (error: any) {
        console.error(error);
        if (error.code === 2000018) {
          Message({ message: '任务提交失败，返回上一步重试提交', theme: 'error' });
        }
      } finally {
        isLoading.value = false;
      }
    };

    const handleSingleZiyanRecycle = (data: any) => {
      // 每次替换上一条
      selections.value.splice(0, selections.value.length, { ...data, __formSingleOp: true });
      operationType.value = OperationActions.RECYCLE;
    };

    expose({
      handleSingleZiyanRecycle,
    });

    const commonTable = () => (
      <CommonLocalTable
        data={tableData.value}
        columns={computedColumns.value}
        changeData={(data) => (tableData.value = data)}
        searchData={searchData}>
        <div class={cssModule['host-operations-toolbar']}>
          <BkButtonGroup>
            <Button onClick={() => handleSwitch(true)} selected={selected.value === 'target'}>
              可{operationMap[operationType.value].label}
            </Button>
            <Button onClick={() => handleSwitch(false)} selected={selected.value === 'untarget'}>
              不可{operationMap[operationType.value].label}
            </Button>
          </BkButtonGroup>
          {computedContent.value}
        </div>
      </CommonLocalTable>
    );

    return () => (
      <>
        <div class={cssModule.host_operations_container}>
          <Dropdown disabled={operationsDisabled.value}>
            {{
              default: () => (
                <Button disabled={operationsDisabled.value}>
                  批量操作
                  <AngleDown class={cssModule.f26}></AngleDown>
                </Button>
              ),
              content: () => (
                <BkDropdownMenu>
                  {Object.entries(operationMap)
                    .filter(([opType]) => opType !== OperationActions.NONE)
                    .map(([opType, opData]) => {
                      const { disabled, tooltips, clickHandler } = getOperationConfig(opType as OperationActions);
                      return withDirectives(
                        <BkDropdownItem
                          onClick={clickHandler}
                          extCls={`more-action-item${disabled ? ' disabled' : ''}`}>
                          批量{opData.label}
                        </BkDropdownItem>,
                        [[bkTooltips, tooltips]],
                      );
                    })}
                  <CopyToClipboard
                    type='dropdown-item'
                    text='复制内网IP'
                    content={selectedRowPrivateIPs.value?.join?.(',')}
                  />
                  <CopyToClipboard
                    type='dropdown-item'
                    text='复制公网IP'
                    content={selectedRowPublicIPs.value?.join?.(',')}
                  />
                </BkDropdownMenu>
              ),
            }}
          </Dropdown>
        </div>

        <Dialog
          isShow={isDialogShow.value}
          quickClose={false}
          title={computedTitle.value}
          ref={dialogRef}
          width={1500}
          closeIcon={!isLoading.value}
          onClosed={handleCancelDialog}>
          {{
            default: () => (
              <Loading loading={isDialogLoading.value}>
                <div class={cssModule['host-operations-main']}>
                  {isZiyanRecycle.value ? (
                    <RecycleFlow
                      ref={recycleFlowRef}
                      ips={hostPrivateIP4s.value}
                      onSelectChange={handleZiyanRecycleSelectChange}>
                      {commonTable()}
                    </RecycleFlow>
                  ) : (
                    <>
                      {computedTips.value && <div class={cssModule['host-operations-tips']}>{computedTips.value}</div>}
                      {commonTable()}
                    </>
                  )}
                </div>
              </Loading>
            ),
            footer: (
              <>
                {isZiyanRecycle.value ? (
                  <>
                    {!recycleFlowRef.value?.isFirstStep?.() && (
                      <Button onClick={() => recycleFlowRef.value.prevStep()} class='mr10'>
                        上一步
                      </Button>
                    )}
                    <Button
                      onClick={
                        recycleFlowRef.value?.isLastStep?.()
                          ? () => handleZiyanRecycleSubmit()
                          : () => recycleFlowRef.value.nextStep()
                      }
                      theme='primary'
                      disabled={
                        recycleFlowRef.value?.isLastStep?.()
                          ? !ziyanRecycleSelected.value.length
                          : isConfirmDisabled.value
                      }
                      loading={isLoading.value}>
                      {recycleFlowRef.value?.isLastStep?.() ? '提交' : '下一步'}
                    </Button>
                  </>
                ) : (
                  <Button
                    onClick={handleConfirm}
                    theme='primary'
                    disabled={isConfirmDisabled.value}
                    loading={isLoading.value}>
                    {operationMap[operationType.value].label}
                  </Button>
                )}
                <Button onClick={handleCancelDialog} class='ml10' disabled={isLoading.value}>
                  取消
                </Button>
              </>
            ),
          }}
        </Dialog>
      </>
    );
  },
});
