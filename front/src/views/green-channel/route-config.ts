import type { RouteRecordRaw } from 'vue-router';
import Meta from '@/router/meta';
import { MENU_GREEN_CHANNEL_MANAGEMENT } from '@/constants/menu-symbol';
import { useGreenChannelQuotaStore } from '@/store/green-channel/quota';

export default [
  {
    name: MENU_GREEN_CHANNEL_MANAGEMENT,
    path: 'green-channel/:module?/:view?',
    component: () => import('./index.vue'),
    beforeEnter: async (to, from, next) => {
      const greenChannelQuotaStore = useGreenChannelQuotaStore();
      try {
        await greenChannelQuotaStore.getGlobalQuota(true);
      } finally {
        next();
      }
    },
    meta: {
      ...new Meta({
        title: '小额绿通',
        activeKey: MENU_GREEN_CHANNEL_MANAGEMENT,
        menu: {
          relative: MENU_GREEN_CHANNEL_MANAGEMENT,
        },
        icon: 'hcm-icon bkhcm-icon-bushu',
      }),
    },
  },
] as RouteRecordRaw[];
