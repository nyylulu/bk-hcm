import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';

const { t } = i18n.global;

export const ticketRoutes: RouteRecordRaw[] = [
  // 兼容老路由
  {
    path: '/service/my-apply',
    redirect: '/service/ticket',
    // meta是必要的，如果不想在侧边栏显示，需要设置notMenu为true
    meta: {
      notMenu: true,
    },
  },
  {
    path: '/service/my-apply/detail',
    redirect: '/service/ticket/detail',
    // meta是必要的，如果不想在侧边栏显示，需要设置notMenu为true
    meta: {
      notMenu: true,
    },
  },
  {
    path: 'ticket',
    name: 'menu_ticket_manage',
    component: () => import('@/views/ticket/entry-srv.vue'),
    meta: {
      activeKey: 'menu_ticket_manage',
      title: t('单据管理'),
      // breadcrumb: [t('服务'), t('我的申请')],
      isShowBreadcrumb: true,
      icon: 'hcm-icon bkhcm-icon-my-apply',
    },
  },
  {
    path: 'ticket/detail',
    name: 'menu_ticket_detail',
    component: () => import('@/views/ticket/children/apply-detail'),
    meta: {
      activeKey: 'menu_ticket_manage',
      notMenu: true,
    },
  },
];

export const ticketRoutesBiz: RouteRecordRaw[] = [
  // 重定向兼容老路由
  {
    path: '/business/applications/detail',
    redirect: '/business/ticket/detail',
    // meta是必要的，如果不想在侧边栏显示，需要设置notMenu为true
    meta: {
      notMenu: true,
    },
  },
  {
    path: '/business/applications/resource-plan/detail',
    redirect: '/business/ticket/resource-plan/detail',
    // meta是必要的，如果不想在侧边栏显示，需要设置notMenu为true
    meta: {
      notMenu: true,
    },
  },
  {
    path: 'ticket',
    name: 'ApplicationsManage',
    component: () => import('@/views/ticket/entry-biz.vue'),
    meta: {
      activeKey: 'applications',
      title: t('单据管理'),
      isShowBreadcrumb: true,
      icon: 'hcm-icon bkhcm-icon-my-apply',
      // notMenu: true,
    },
  },
  {
    path: 'ticket/detail',
    name: '申请单据详情',
    component: () => import('@/views/ticket/children/apply-detail'),
    meta: {
      activeKey: 'applications',
      notMenu: true,
    },
  },
  // 资源管理下 单据管理 tab 资源预测详情
  {
    path: 'ticket/resource-plan/detail',
    name: 'BizInvoiceResourceDetail',
    component: () => import('@/views/ticket/children/resource-plan/detail'),
    meta: {
      activeKey: 'applications',
      notMenu: true,
    },
  },
];

export default ticketRoutes;
