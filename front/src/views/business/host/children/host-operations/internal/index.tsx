import { PropType, computed, defineComponent, ref, toRefs, useTemplateRef, withDirectives } from 'vue';
import { useRouter } from 'vue-router';
import cssModule from '../index.module.scss';

import { useZiyanScrStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useBatchOperation from './use-batch-operation';
import { operationMap as defaultOperationMap, OperationMapItem } from '../index';
import cvmStatusBaseColumns from '../constants/cvm-status-base-columns';

import { Button, Dialog, Dropdown, Message, bkTooltips } from 'bkui-vue';
import { BkDropdownItem, BkDropdownMenu } from 'bkui-vue/lib/dropdown';
import { BkButtonGroup } from 'bkui-vue/lib/button';
import { AngleDown } from 'bkui-vue/lib/icon';
import RecycleFlow from './recycle-flow.vue';
import CommonLocalTable from '@/components/LocalTable';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import HostBatchResetDialog from './host-batch-reset-dialog/index.vue';
import CvmStatusCollapseTable from '../children/cvm-status-collapse-table.vue';
import CvmStatusTable from '../children/cvm-status-table.vue';
import MoaVerifyBtn from '@/components/moa-verify/moa-verify-btn.vue';

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
  [OperationActions.RECYCLE]: {
    ...defaultOperationMap[OperationActions.RECYCLE],
    // 鉴权参数
    authId: 'biz_iaas_resource_delete',
    actionName: 'biz_iaas_resource_delete',
  },
  [OperationActions.RESET]: {
    label: '重装',
    labelEn: 'Reset',
    disabledStatus: [] as string[],
    loading: false,
    // 鉴权参数
    authId: 'biz_iaas_resource_operate',
    actionName: 'biz_iaas_resource_operate',
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
      computedColumns,
      tableData,
      targetHost,
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
      moaVerifyRef,
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
        hostBatchResetDialogRef.value.show(selections.value.map((v) => v.id));
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

    // 主机重装
    const hostBatchResetDialogRef = useTemplateRef<typeof HostBatchResetDialog>('host-batch-reset-dialog');

    // 开/关机、重启 MOA校验
    const moaVerifyPromptPayload = computed(() => {
      const operateName = operationMap[operationType.value].label;
      const operateNameEn = operationMap[operationType.value].labelEn;
      const operateNum = tableData.value.length;

      return {
        zh: {
          title: `HCM-主机${operateName}确认`,
          desc: `您正在对${operateNum}台主机执行${operateName}，是否同意本次操作？`,
        },
        en: {
          title: `HCM-Host ${operateNameEn} Verification`,
          desc: `${operateNameEn} OS on ${operateNum} host(s). Do you agree to this operation?`,
        },
      };
    });

    const footerRef = useTemplateRef<HTMLElement>('footer');

    expose({
      handleSingleZiyanRecycle,
      hostBatchResetDialogRef,
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

        {/* 开机、关机、重启、回收操作 dialog */}
        <Dialog
          isShow={isDialogShow.value}
          quickClose={false}
          title={computedTitle.value}
          ref={dialogRef}
          width={1500}
          closeIcon={!isLoading.value}
          onClosed={handleCancelDialog}
          renderDirective='if'>
          {{
            default: () => (
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
                    {/* 自研云主机操作 */}
                    {isZiyanOnly.value ? (
                      <>
                        <div class={[cssModule['host-operations-toolbar'], 'mb12']}>
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
                        {selected.value === 'target' ? (
                          <CvmStatusTable list={tableData.value} columns={cvmStatusBaseColumns} />
                        ) : (
                          <CvmStatusCollapseTable list={tableData.value} />
                        )}
                      </>
                    ) : (
                      // 公有云主机操作
                      commonTable()
                    )}
                  </>
                )}
              </div>
            ),
            footer: (
              <div class={cssModule.footer} ref='footer'>
                {(function () {
                  // 自研云回收
                  if (isZiyanRecycle.value) {
                    return (
                      <>
                        {!recycleFlowRef.value?.isFirstStep?.() && (
                          <Button class={cssModule.button} onClick={() => recycleFlowRef.value.prevStep()}>
                            上一步
                          </Button>
                        )}
                        <Button
                          class={cssModule.button}
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
                    );
                  }
                  // 自研云开/关机、重启
                  if (isZiyanOnly.value) {
                    return (
                      <>
                        {targetHost.value.length > 0 && (
                          <MoaVerifyBtn
                            ref='moa-verify'
                            class={cssModule['moa-verify-btn']}
                            verifyText='MOA校验'
                            promptPayload={moaVerifyPromptPayload.value}
                            boundary={footerRef.value}
                          />
                        )}
                        <Button
                          class={cssModule.button}
                          onClick={handleConfirm}
                          theme='primary'
                          disabled={moaVerifyRef.value?.verifyResult.button_type !== 'confirm'}
                          loading={isLoading.value}>
                          {operationMap[operationType.value].label}
                        </Button>
                      </>
                    );
                  }
                  // 公有云开/关机、重启、回收
                  return (
                    <Button
                      class={cssModule.button}
                      onClick={handleConfirm}
                      theme='primary'
                      disabled={isConfirmDisabled.value}
                      loading={isLoading.value}>
                      {operationMap[operationType.value].label}
                    </Button>
                  );
                })()}
                <Button onClick={handleCancelDialog} class={cssModule.button} disabled={isLoading.value}>
                  取消
                </Button>
              </div>
            ),
          }}
        </Dialog>

        {/* 批量重装dialog */}
        <HostBatchResetDialog ref='host-batch-reset-dialog' />
      </>
    );
  },
});
