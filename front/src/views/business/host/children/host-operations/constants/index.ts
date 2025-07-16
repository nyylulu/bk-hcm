import i18n from '@/language/i18n';
import { ConstantMapRecord } from '@/typings';
import { ImageType } from '../typings';

const { t } = i18n.global;

export const OPERATE_STATUS_MAP: ConstantMapRecord = {
  0: t('正常'),
  1: t('不是主备负责人，无权限进行该操作'),
  2: t('不在空闲机模块，不可重装'),
  3: t('云服务器未处于关机状态'),
  4: t('云服务器未处于开机状态'),
  5: t('物理机不支持操作'),
};

// 物理机
export const IDC_SVC_SOURCE_TYPE_IDS = ['1', '2', '3'];
// 虚拟机
export const CVM_SVC_SOURCE_TYPE_IDS = ['4', '5'];

export const IMAGE_TYPE_NAME = {
  [ImageType.PUBLIC_IMAGE]: t('公共镜像'),
  [ImageType.PRIVATE_IMAGE]: t('私有镜像'),
};
