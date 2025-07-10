import type { RouteRecordRaw } from 'vue-router';

const ziyanScr: RouteRecordRaw[] = [
  {
    path: '/ziyanScr',
    children: [
      {
        path: '/ziyanScr/hostInventory',
        component: () => import('@/views/ziyanScr/hostInventory/index'),
        meta: {
          title: '主机库存',
          activeKey: 'zzkc',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
      {
        path: '/ziyanScr/hostApplication',
        component: () => import('@/views/ziyanScr/hostApplication'),
        meta: {
          title: '主机申领',
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
        path: '/ziyanScr/hostApplication/modify',
        name: '修改主机申请',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-modify/index.vue'),
        meta: {
          activeKey: 'zjsq',
          notMenu: true,
        },
      },
      {
        path: '/ziyanScr/hostRecycling',
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
          {
            path: 'docDetail',
            name: 'docDetail',
            component: () => import('@/views/ziyanScr/host-recycle/bill-detail'),
            meta: {
              activeKey: 'zjhs',
              breadcrumb: ['资源', '主机'],
            },
          },
        ],
        meta: {
          title: '主机回收',
          activeKey: 'zjhs',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
    ],
    meta: {
      groupTitle: '资源',
    },
  },
  {
    path: '/ziyanScr',
    children: [
      {
        path: '/ziyanScr/jfcc',
        component: () => import('@/views/ziyanScr/recycle-server-room'),
        children: [],
        meta: {
          title: '机房裁撤',
          activeKey: 'jfcc',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
        },
      },
    ],
    meta: {
      groupTitle: '服务',
    },
  },
  {
    path: '/ziyanScr',
    children: [
      {
        path: '/ziyanScr/cvm-model',
        component: () => import('@/views/ziyanScr/cvm-model'),
        children: [],
        meta: {
          title: 'CVM机型',
          activeKey: 'cvmjx',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
          checkAuth: 'ziyan_cvm_type_find',
        },
      },
      {
        path: '/ziyanScr/cvmzw',
        component: () => import('@/views/ziyanScr/cvm-web'),
        children: [],
        meta: {
          title: 'CVM子网',
          activeKey: 'cvmzw',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
          checkAuth: 'ziyan_cvm_subnet_find',
        },
      },
      {
        path: '/ziyanScr/resource-manage',
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
          title: '资源上下架',
          activeKey: 'scr-resource-manage',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
          checkAuth: 'ziyan_res_shelves_find',
        },
      },
      {
        path: '/ziyanScr/cvmsc',
        component: () => import('@/views/ziyanScr/cvm-produce'),
        children: [],
        meta: {
          title: 'CVM生产',
          activeKey: 'cvmsc',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host',
          checkAuth: 'ziyan_cvm_create_find',
        },
      },
    ],
    meta: {
      groupTitle: '管理',
    },
  },
];

export default ziyanScr;
