import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_ROLLING_SERVER_MANAGEMENT } from '@/constants/menu-symbol';

export default [
  {
    name: MENU_ROLLING_SERVER_MANAGEMENT,
    path: 'rolling-server/:module?/:view?',
    component: () => import('./index.vue'),
    meta: {
      ...new Meta({
        title: '滚服管理',
        // 视图权限，暂未实现
        auth: {
          view: { type: 'biz_access' },
        },
        activeKey: MENU_ROLLING_SERVER_MANAGEMENT,
        menu: {
          relative: MENU_ROLLING_SERVER_MANAGEMENT,
        },
        icon: 'hcm-icon bkhcm-icon-bushu',
      }),
    },
  },
] as RouteRecordRaw[];
