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
        component: () => import('@/views/ziyanScr/hostApplication'),
        meta: {
          activeKey: 'zjsq',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/ziyanScr/hostApplication/detail/:id',
        name: 'host-application-detail',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-detail/index'),
        meta: {
          activeKey: 'zjsq',
          notMenu: true,
        },
      },
      {
        path: '/ziyanScr/hostApplication/apply',
        name: '提交主机申请',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-form/index'),
        meta: {
          activeKey: 'zjsq',
          notMenu: true,
        },
      },
      {
        path: '/ziyanScr/hostRecycling',
        name: '主机回收',
        children: [
          {
            path: '',
            name: 'hostRecycle',
            component: () => import('@/views/ziyanScr/host-recycle'),
            meta: {
              activeKey: 'zjhs',
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'resources',
            name: 'resources',
            component: () => import('@/views/ziyanScr/RecyclingResources'),
            meta: {
              activeKey: 'zjhs',
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'preDetail',
            name: 'PreDetail',
            component: () => import('@/views/ziyanScr/host-recycle/pre-details'),
            meta: {
              activeKey: 'zjhs',
              breadcrumb: ['资源', '主机'],
            },
          },
        ],
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
        component: () => import('@/views/ziyanScr/recycle-server-room'),
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
        path: '/ziyanScr/cvm-model',
        component: () => import('@/views/ziyanScr/cvm-model'),
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
        component: () => import('@/views/ziyanScr/cvm-web'),
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
        children: [
          {
            path: '',
            name: 'resourceManage',
            component: () => import('@/views/ziyanScr/resource-manage'),
            meta: {
              activeKey: 'scr-resource-manage',
            },
          },
          {
            path: 'detail/:id',
            name: 'scrResourceManageDetail',
            component: () => import('@/views/ziyanScr/resource-manage/detail'),
            props(route) {
              return { ...route.params, ...route.query };
            },
            meta: {
              activeKey: 'scr-resource-manage',
            },
          },
          {
            path: 'create',
            name: 'scrResourceManageCreate',
            component: () => import('@/views/ziyanScr/resource-manage/create'),
            props(route) {
              return { ...route.query };
            },
            meta: {
              activeKey: 'scr-resource-manage',
            },
          },
        ],
        meta: {
          activeKey: 'scr-resource-manage',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/ziyanScr/cvmsc',
        name: 'CVM生产',
        component: () => import('@/views/ziyanScr/cvm-produce'),
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
