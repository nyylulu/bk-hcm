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
import { OperationActions, operationMap } from '../index';
import RecycleFlow from './recycle-flow.vue';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';

export { OperationActions, operationMap } from '../index';

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

    const operationDisabledTips = (type: string) => {
      const isRecycle = type === OperationActions.RECYCLE;
      if (isMix.value) {
        return {
          content: '腾讯云自研云和公有云的主机，不支持同时操作',
          disabled: false,
        };
      }
      if (isZiyanOnly.value) {
        return {
          content: isRecycle ? '' : '暂不支持',
          disabled: isRecycle,
        };
      }
      return {
        content: '',
        disabled: true,
      };
    };

    const handleClickMenu = (type: OperationActions) => {
      if (!operationDisabledTips(type).disabled) {
        return;
      }
      operationType.value = type;
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
                      return withDirectives(
                        <BkDropdownItem
                          onClick={() => {
                            if (!authVerifyData?.value?.permissionAction?.biz_iaas_resource_delete) {
                              handleAuth('biz_iaas_resource_delete');
                              globalPermissionDialog.setShow(true);
                            } else handleClickMenu(opType as OperationActions);
                          }}
                          extCls={`more-action-item${
                            !operationDisabledTips(opType as OperationActions).disabled ||
                            (opType === OperationActions.RECYCLE &&
                              !authVerifyData?.value?.permissionAction?.biz_iaas_resource_delete)
                              ? ' disabled'
                              : ''
                          }`}>
                          批量{opData.label}
                        </BkDropdownItem>,
                        [[bkTooltips, operationDisabledTips(opType as OperationActions)]],
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
                          ? !ziyanRecycleSelected.value.length ||
                            recycleFlowRef.value?.isRollingServerCpuCoreExceedByResPool
                          : isConfirmDisabled.value
                      }
                      loading={isLoading.value}
                      v-bk-tooltips={{
                        content: '资源池业务下，选择为“滚服项目”的核数，不能超过全平台应该退还给公司的额度',
                        disabled: !recycleFlowRef.value?.isRollingServerCpuCoreExceedByResPool,
                      }}>
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
