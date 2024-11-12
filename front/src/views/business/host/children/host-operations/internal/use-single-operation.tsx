import { ref } from 'vue';
import { CLOUD_HOST_STATUS, VendorEnum } from '@/common/constant';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';
import { operationMap, OperationActions } from './index';

const useSingleOperation = ({ customOperate }: { customOperate: Function }) => {
  const { authVerifyData, handleAuth } = useVerify();
  const globalPermissionDialog = useGlobalPermissionDialog();

  const currentOperateRowIndex = ref(-1);

  const handleClickMenu = (type: OperationActions, data: any) => {
    if (getOperationConfig(type, data).disabled) {
      return;
    }

    customOperate(type, data);
  };

  const getOperationConfig = (type: OperationActions, data: any) => {
    // 点击事件（值缺省时，为默认点击事件）
    const clickHandler = () => handleClickMenu(type, data);

    const statusDisabled = operationMap[type].disabledStatus.includes(data.status);

    const isNotZiyanReset = data.vendor !== VendorEnum.ZIYAN && type === OperationActions.RESET;

    if (isNotZiyanReset) {
      return {
        disabled: true,
        tooltips: { content: '暂不支持', disabled: false },
        clickHandler,
      };
    }

    if (statusDisabled) {
      return {
        disabled: true,
        tooltips: { content: `当前主机处于 ${CLOUD_HOST_STATUS[data.status]} 状态`, disabled: false },
        clickHandler,
      };
    }

    // 预鉴权
    const { authId, actionName } = operationMap[type];
    const noPermission = !authVerifyData?.value?.permissionAction?.[authId];
    if (authId && actionName && noPermission) {
      return {
        disabled: false,
        noPermission: true,
        tooltips: { disabled: true },
        clickHandler: () => {
          handleAuth(actionName);
          globalPermissionDialog.setShow(true);
        },
      };
    }

    return { disabled: false, tooltips: { disabled: true }, clickHandler };
  };

  return {
    currentOperateRowIndex,
    getOperationConfig,
  };
};

export default useSingleOperation;
