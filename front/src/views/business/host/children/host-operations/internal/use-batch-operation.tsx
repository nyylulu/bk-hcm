import { computed, Ref, useTemplateRef, watch } from 'vue';
import { Message } from 'bkui-vue';
import { useZiyanScrStore } from '@/store';
import { VendorEnum } from '@/common/constant';
import { ResourceTypeEnum } from '@/common/resource-constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useBusinessStore } from '@/store/business';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';
import { OperationActions } from './index';
import defaultUseBatchOperation, { type Params } from '../use-batch-operation';
import { CvmOperateType, useCvmOperateStore } from '@/store/cvm-operate';
import { getPrivateIPs, getPublicIPs } from '@/utils';

const useBatchOperation = ({ selections, onFinished }: Params) => {
  const {
    operationType: defaultOperationType,
    isDialogShow: defaultIsDialogShow,
    isConfirmDisabled,
    operationsDisabled,
    computedTitle,
    computedTips,
    computedContent,
    isLoading,
    selected,
    isDialogLoading,
    baseColumns,
    recycleColumns,
    tableData,
    targetHost,
    unTargetHost,
    searchData,
    selectedRowPrivateIPs,
    selectedRowPublicIPs,
    getDiskNumByCvmIds,
    handleSwitch,
    handleConfirm: defaultHandleConfirm,
    handleCancelDialog,
  } = defaultUseBatchOperation({
    selections,
    onFinished,
  });

  const operationType: Ref<OperationActions> = defaultOperationType;

  const { getBizsId } = useWhereAmI();
  const scrStore = useZiyanScrStore();
  const businessStore = useBusinessStore();
  const cvmOperateStore = useCvmOperateStore();

  // 重装操作不集成到 host-operations 组件中
  const isDialogShow = computed(() => defaultIsDialogShow.value && OperationActions.RESET !== operationType.value);

  const vendorSet = computed(() => {
    const vendors = selections.value.map((item) => item.vendor);
    return new Set(vendors);
  });

  const isMix = computed(() => {
    return vendorSet.value.size > 1 && vendorSet.value.has(VendorEnum.ZIYAN);
  });

  const isZiyanOnly = computed(() => vendorSet.value.size === 1 && vendorSet.value.has(VendorEnum.ZIYAN));
  const isRecycle = computed(() => operationType.value === OperationActions.RECYCLE);
  const isZiyanRecycle = computed(() => isRecycle.value && isZiyanOnly.value);

  const computedColumns = computed(() => {
    const columns = baseColumns.value.slice();

    if (isZiyanRecycle.value) {
      return getZiyanRecycleColumn(columns);
    }

    if (isRecycle.value) {
      return columns.concat(recycleColumns.value);
    }

    return columns;
  });

  const hostPrivateIP4s = computed(() =>
    selections.value.reduce((acc, cur) => {
      return acc.concat(cur?.private_ipv4_addresses || []);
    }, []),
  );

  // 只需要preview可回收的主机
  const previewHostIps = computed(() => targetHost.value?.flatMap((cur) => cur.private_ipv4_addresses) || []);

  const getZiyanRecycleColumn = (defaultColumns: typeof baseColumns.value) => {
    const columns = defaultColumns.slice();
    columns.unshift({
      field: '_asset_id',
      label: '固资号',
    });

    columns.push({
      field: '_topo_module',
      label: '所属模块',
    });

    if (selected.value === 'untarget') {
      columns.push({
        field: '_message',
        label: '不可回收原因',
      });
    }

    columns.push(
      {
        field: '_operator',
        label: '主负责人',
      },
      {
        field: '_bak_operator',
        label: '备份负责人',
      },
      {
        field: '_input_time',
        label: '入库时间',
      },
    );

    return columns;
  };

  const getRecyclableStatus = async () => {
    const result = await scrStore.getRecyclableHosts({ ips: hostPrivateIP4s.value });
    const recycleStatusList = result?.data?.info ?? [];

    // 先清空在公共hook中设置的数据，在这里重新赋值
    targetHost.value = [];
    unTargetHost.value = [];

    for (const host of selections.value) {
      const found = recycleStatusList.find((item: any) => host.private_ipv4_addresses.includes(item.ip));
      const newHost = {
        ...host,
        _asset_id: found?.asset_id,
        _topo_module: found?.topo_module,
        _operator: found?.operator,
        _bak_operator: found?.bak_operator,
        _input_time: found?.input_time,
        _message: found?.message,
      };
      if (found?.recyclable) {
        targetHost.value.push(newHost);
      } else {
        unTargetHost.value.push(newHost);
      }
    }
  };

  const getZiyanCvmOperateStatus = async (type: OperationActions) => {
    isDialogLoading.value = true;

    // 先清空在公共hook中设置的数据
    targetHost.value = [];
    unTargetHost.value = [];

    try {
      const ids = selections.value.map((v) => v.id);
      const operate_type = type as CvmOperateType;
      const list = (await cvmOperateStore.getCvmListOperateStatus({ ids, operate_type })).map((item) => ({
        ...item,
        private_ip_address: getPrivateIPs(item),
        public_ip_address: getPublicIPs(item),
      }));

      // operate_status 为 0，push 到 targetHost.value 中；其他状态值，push 到 unTargetHost.value 中
      targetHost.value = list.filter((v) => v.operate_status === 0);
      unTargetHost.value = list.filter((v) => v.operate_status !== 0);
    } finally {
      isDialogLoading.value = false;
    }
  };

  watch(operationType, async (type) => {
    if (isZiyanRecycle.value) {
      await getRecyclableStatus();
    }

    // 自研云主机操作：开/关机、重启
    if (isZiyanOnly.value && [OperationActions.START, OperationActions.STOP, OperationActions.REBOOT].includes(type)) {
      await getZiyanCvmOperateStatus(type);
    }

    // 公有云回收
    if (!isZiyanOnly.value && isRecycle.value) {
      await getDiskNumByCvmIds();
    }

    isConfirmDisabled.value = targetHost.value.length === 0;
    handleSwitch(targetHost.value.length > 0);
  });

  const moaVerifyRef = useTemplateRef<any>('moa-verify');
  const handleConfirm = async () => {
    if (isZiyanOnly.value) {
      const hostIds = targetHost.value.map((v) => v.id);
      const session_id = moaVerifyRef.value?.verifyResult.session_id;

      try {
        isLoading.value = true;

        const result = await businessStore.cvmOperateAsync(operationType.value, { ids: hostIds, session_id });

        Message({ theme: 'success', message: '操作成功' });

        // 跳转至新任务详情页
        routerAction.redirect({
          name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
          params: { resourceType: ResourceTypeEnum.CVM, id: result.data.task_management_id },
          query: { bizs: getBizsId() },
        });
        operationType.value = OperationActions.NONE;
      } catch (error: any) {
        if (error.code === 2000019) {
          // MOA校验过期
          Message({ theme: 'error', message: 'MOA校验过期，请重新发起校验后操作' });
          moaVerifyRef.value?.resetVerifyResult();
        } else {
          Message({ theme: 'error', message: error.message });
        }
      } finally {
        isLoading.value = false;
      }
    } else {
      defaultHandleConfirm();
    }
  };

  return {
    operationType,
    isDialogShow,
    computedColumns,
    computedTitle,
    computedTips,
    computedContent,
    operationsDisabled,
    isConfirmDisabled,
    isMix,
    isZiyanOnly,
    isZiyanRecycle,
    hostPrivateIP4s,
    previewHostIps,
    isLoading,
    tableData,
    targetHost,
    selected,
    isDialogLoading,
    searchData,
    selectedRowPrivateIPs,
    selectedRowPublicIPs,
    moaVerifyRef,
    handleSwitch,
    handleConfirm,
    handleCancelDialog,
  };
};

export default useBatchOperation;
