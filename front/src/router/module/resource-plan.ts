import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';

const { t } = i18n.global;

const resourcePlanMenus: RouteRecordRaw[] = [
  // {
  //   path: '/resource-plan/manage',
  //   name: t('单据管理'),
  //   component: () => import('@/views/resource-plan/manage/index'),
  //   meta: {
  //     activeKey: 'planManage',
  //     icon: 'hcm-icon bkhcm-icon-template-orchestration',
  //   },
  // },
  {
    path: '/resource-plan/detail',
    component: () => import('@/views/resource-plan/detail/index'),
    meta: {
      activeKey: 'planlist',
      notMenu: true,
    },
  },
  {
    path: '/resource-plan/list',
    name: t('资源预测'),
    component: () => import('@/views/resource-plan/list/index'),
    meta: {
      activeKey: 'planlist',
      icon: 'hcm-icon bkhcm-icon-template-orchestration',
    },
  },
  {
    path: '/resource-plan/add',
    component: () => import('@/views/resource-plan/add/index'),
    meta: {
      activeKey: 'planlist',
      notMenu: true,
    },
  },
];

export default resourcePlanMenus;
