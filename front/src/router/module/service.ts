import type { RouteRecordRaw } from 'vue-router';
import i18n from '@/language/i18n';

const { t } = i18n.global;

const serviceMenus: RouteRecordRaw[] = [
  {
    path: '/service',
    children: [
      {
        path: '/service/service-apply',
        name: 'serviceApply',
        component: () => import('@/views/service/service-apply/index.vue'),
        meta: {
          title: t('服务申请'),
          activeKey: 'serviceApply',
          // breadcrumb: [t('服务'), t('服务申请')],
          notMenu: true,
          isShowBreadcrumb: true,
        },
      },
      {
        path: '/service/my-apply',
        name: 'myApply',
        component: () => import('@/views/service/apply-list/index'),
        // component: () => import('@/views/service/my-apply/index.vue'),
        meta: {
          activeKey: 'myApply',
          title: t('单据管理'),
          // breadcrumb: [t('服务'), t('我的申请')],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-my-apply',
        },
      },
      // 单据管理 tab 资源预测详情
      {
        path: '/service/my-apply/resource-plan/detail',
        name: 'OpInvoiceResourceDetail',
        component: () => import('@/views/resource-plan/invoice-manage/detail/index'),
        meta: {
          activeKey: 'opInvoiceResourceDetail',
          notMenu: true,
        },
      },
      {
        path: '/service/my-apply/detail',
        name: 'serviceMyApplyDetail',
        component: () => import('@/views/service/apply-detail/index'),
        meta: {
          activeKey: 'myApply',
          notMenu: true,
        },
      },
      {
        path: '/service/my-approval',
        name: t('我的审批'),
        component: () => import('@/views/service/my-approval/page'),
        meta: {
          // breadcrumb: [t('服务'), t('我的审批')],
          isShowBreadcrumb: true,
          notMenu: true,
        },
      },
      {
        path: '/service/dissolve',
        component: () => import('@/views/ziyanScr/recycle-server-room'),
        children: [],
        meta: {
          title: t('机房裁撤'),
          activeKey: 'dissolve',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-dissolve',
        },
      },
    ],
    meta: {
      groupTitle: '资源',
    },
  },
  {
    path: '/service',
    children: [
      {
        path: '/service/resource-plan',
        name: 'opResourcePlan',
        component: () => import('@/views/resource-plan/resource-manage/list'),
        meta: {
          activeKey: 'opResourcePlan',
          title: t('资源预测'),
          isShowBreadcrumb: true,
          icon: '',
        },
      },
      {
        path: '/service/resource-plan/detail',
        name: 'opResourcePlanDetail',
        component: () => import('@/views/resource-plan/resource-manage/detail'),
        meta: {
          activeKey: 'opResourcePlanDetail',
          notMenu: true,
        },
      },
      {
        path: '/service/hostInventory',
        component: () => import('@/views/ziyanScr/hostInventory/index'),
        meta: {
          title: t('主机库存'),
          activeKey: 'inventory',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host-inventory',
          checkAuth: 'ziyan_resource_inventory_find',
        },
      },
      {
        path: '/service/hostApplication',
        component: () => import('@/views/ziyanScr/hostApplication'),
        name: '主机申领',
        meta: {
          title: t('主机申领'),
          activeKey: 'apply',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host-application',
          checkAuth: 'ziyan_resource_create',
        },
      },
      {
        path: '/service/hostApplication/detail/:id',
        name: 'host-application-detail',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-detail/index'),
        meta: {
          activeKey: 'apply',
          notMenu: true,
        },
      },
      {
        path: '/service/hostApplication/apply',
        name: '提交主机申请',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-form/index'),
        meta: {
          activeKey: 'apply',
          notMenu: true,
        },
      },
      {
        path: '/service/hostApplication/modify',
        name: '修改主机申请',
        component: () => import('@/views/ziyanScr/hostApplication/components/application-modify/index'),
        meta: {
          activeKey: 'apply',
          notMenu: true,
        },
      },
      {
        path: '/service/hostRecycling',
        name: '主机回收',
        children: [
          {
            path: '',
            name: 'hostRecycle',
            component: () => import('@/views/ziyanScr/host-recycle'),
            meta: {
              activeKey: 'recovery',
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'resources',
            name: 'resources',
            component: () => import('@/views/ziyanScr/RecyclingResources'),
            meta: {
              activeKey: 'recovery',
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'preDetail',
            name: 'PreDetail',
            component: () => import('@/views/ziyanScr/host-recycle/pre-details'),
            meta: {
              activeKey: 'recovery',
              breadcrumb: ['资源', '主机'],
            },
          },
          {
            path: 'docDetail',
            name: 'docDetail',
            component: () => import('@/views/ziyanScr/host-recycle/bill-detail'),
            meta: {
              activeKey: 'recovery',
              breadcrumb: ['资源', '主机'],
            },
          },
        ],
        meta: {
          activeKey: 'recovery',
          title: t('主机回收'),
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-host-recycle',
          checkAuth: 'ziyan_resource_recycle',
        },
      },
      {
        path: '/service/cvm-model',
        component: () => import('@/views/ziyanScr/cvm-model'),
        name: 'CVM机型',
        children: [],
        meta: {
          activeKey: 'model',
          title: t('CVM机型'),
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-cvm-type',
          checkAuth: 'ziyan_cvm_type_find',
        },
      },
      {
        path: '/service/cvm-subnet',
        name: 'CVM子网',
        component: () => import('@/views/ziyanScr/cvm-web'),
        children: [],
        meta: {
          title: t('CVM子网'),
          activeKey: 'subnet',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-subnet',
          checkAuth: 'ziyan_cvm_subnet_find',
        },
      },
      {
        path: '/service/resource-manage',
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
          title: t('资源上下架'),
          activeKey: 'scr-resource-manage',
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-res-shelves',
          checkAuth: 'ziyan_res_shelves_find',
        },
      },
      {
        path: '/service/cvm-produce',
        name: 'CVM生产',
        component: () => import('@/views/ziyanScr/cvm-produce'),
        children: [],
        meta: {
          title: t('CVM生产'),
          activeKey: 'produce',
          breadcrumb: ['资源', '主机'],
          isShowBreadcrumb: true,
          icon: 'hcm-icon bkhcm-icon-cvm-produce',
          checkAuth: 'ziyan_cvm_create_find',
        },
      },
    ],
    meta: {
      groupTitle: '管理',
    },
  },
];

export default serviceMenus;
