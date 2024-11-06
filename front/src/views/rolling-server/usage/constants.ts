import { AppliedType, ReturnedWay } from '@/store';

export const APPLIED_TYPE_NAME = {
  [AppliedType.NORMAL]: '普通申请',
  [AppliedType.CVM_PRODUCT]: '管理员cvm生产',
  [AppliedType.RESOURCE_POOL]: '资源池申请',
};

export const RETURNED_WAY_NAME = {
  [ReturnedWay.CRP]: '通过crp退还',
  [ReturnedWay.RESOURCE_POOL]: '通过转移到资源池退还',
};
