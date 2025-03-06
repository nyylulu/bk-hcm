import { ResourcesDemandsStatus } from '@/typings/resourcePlan';

export const RESOURCE_DEMANDS_STATUS_NAME = {
  [ResourcesDemandsStatus.CAN_APPLY]: '可申领',
  [ResourcesDemandsStatus.NOT_READY]: '未到申领时间',
  [ResourcesDemandsStatus.EXPIRED]: '已过期',
  [ResourcesDemandsStatus.SPENT_ALL]: '额度用尽',
  [ResourcesDemandsStatus.LOCKED]: '变更中',
};

export const RESOURCE_DEMANDS_STATUS_CLASSES = {
  [ResourcesDemandsStatus.CAN_APPLY]: 'c-success',
  [ResourcesDemandsStatus.NOT_READY]: 'c-info',
  [ResourcesDemandsStatus.EXPIRED]: 'c-info',
  [ResourcesDemandsStatus.SPENT_ALL]: 'c-info',
  [ResourcesDemandsStatus.LOCKED]: 'c-warning',
};
