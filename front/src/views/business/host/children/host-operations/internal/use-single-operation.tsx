import { ref } from 'vue';
import { CLOUD_HOST_STATUS, VendorEnum } from '@/common/constant';
import { operationMap, OperationActions } from '@/views/business/host/children/host-operations';

const useSingleOperation = () => {
  const currentOperateRowIndex = ref(-1);

  const isOperateDisabled = (type: OperationActions, data: any) => {
    // 自研云非回收操作都先disabled
    if (data.vendor === VendorEnum.ZIYAN && type !== OperationActions.RECYCLE) {
      return true;
    }
    return operationMap[type].disabledStatus.includes(data.status);
  };

  const getMenuTooltipsOption = (
    type: OperationActions,
    data: { vendor: VendorEnum; status: keyof typeof CLOUD_HOST_STATUS },
  ) => {
    return {
      content: data.vendor !== VendorEnum.ZIYAN ? `当前主机处于 ${CLOUD_HOST_STATUS[data.status]} 状态` : '暂不支持',
      disabled: !isOperateDisabled(type, data),
    };
  };

  return {
    currentOperateRowIndex,
    isOperateDisabled,
    getMenuTooltipsOption,
  };
};

export default useSingleOperation;
