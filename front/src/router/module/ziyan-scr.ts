import type { RouteRecordRaw } from 'vue-router';

const ziyanScr: RouteRecordRaw[] = [
  {
    path: '/ziyanScr',
    children: [
      {
        path: '/ziyanScr/hostInventory',
        name: '主机库存',
        component: () => import('@/views/ziyanScr/hostInventory/index'),
        meta: {
          activeKey: 'zzkc',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/ziyanScr/hostApplication',
        name: '主机申请',
        component: () => import('@/views/ziyanScr/hostApplication/index'),
        meta: {
          activeKey: 'zjsq',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/ziyanScr/hostRecycling',
        name: '主机回收',
        component: () => import('@/views/ziyanScr/host-recycle'),
        children: [],
        meta: {
          activeKey: 'zjhs',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
    ],
    meta: {
      menuName: '资源',
    },
  },
  {
    path: '/ziyanScr',
    children: [
      {
        path: '/ziyanScr/jfcc',

        name: '机房裁撤',
        children: [],
        meta: {
          activeKey: 'jfcc',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
    ],
    meta: {
      menuName: '服务',
    },
  },
  {
    path: '/ziyanScr',
    children: [
      {
        path: '/ziyanScr/cvmjx',

        name: 'CVM机型',
        children: [],
        meta: {
          activeKey: 'cvmjx',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/ziyanScr/cvmzw',

        name: 'CVM子网',
        children: [],
        meta: {
          activeKey: 'cvmzw',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/ziyanScr/resource-manage',
        name: '资源上下架',
        component: () => import('@/views/ziyanScr/resource-manage'),
        children: [],
        meta: {
          activeKey: 'scr-resource-manage',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/ziyanScr/cvmsc',

        name: 'CVM生产',
        children: [],
        meta: {
          activeKey: 'cvmsc',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
    ],
    meta: {
      menuName: '管理',
    },
  },
];

export default ziyanScr;
