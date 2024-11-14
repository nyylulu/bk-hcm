import i18n from '@/language/i18n';
import { ConstantMapRecord } from '@/typings';
import { ImageType } from './typings';

const { t } = i18n.global;

export const RESET_STATUS_MAP: ConstantMapRecord = {
  1: t('不是主备负责人，无权限进行该操作'),
  2: t('不在空闲机模块，不可重装'),
  3: t('CC运营状态不在“重装中”，不可重装'),
};

// 物理机
export const IDC_SVC_SOURCE_TYPE_IDS = ['1', '2', '3'];
// 虚拟机
export const CVM_SVC_SOURCE_TYPE_IDS = ['4', '5'];

export const IMAGE_TYPE_NAME = {
  [ImageType.PRIVATE_IMAGE]: t('私有镜像'),
  [ImageType.PUBLIC_IMAGE]: t('公共镜像'),
};
