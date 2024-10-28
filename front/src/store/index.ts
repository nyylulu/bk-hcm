import { useUserStore } from './user';
import { useBusinessGlobalStore } from './business-global';

export const preload = async () => {
  const { userInfo } = useUserStore();
  const { getFullBusiness } = useBusinessGlobalStore();

  return Promise.all([userInfo(), getFullBusiness()]);
};

export * from './staff';
export * from './user';
export * from './account';
export * from './departments';
export * from './business';
export * from './ziyanScr';
export * from './resource';
export * from './resourcePlan';
export * from './common';
export * from './host';
export * from './scheme';
export * from './loadbalancer';
export * from './ziyanScr';
export * from './rolling-server';
