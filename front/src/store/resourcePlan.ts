import http from '@/http';
import { defineStore } from 'pinia';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useResourcePlanStore = defineStore({
  id: 'resourcePlanStore',
  state: () => ({}),
  actions: {
    getDiskTypes() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/disk_type/list`);
    },
  },
});
