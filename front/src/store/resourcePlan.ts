import http from '@/http';
import { defineStore } from 'pinia';
import {
  IListTicketsParam,
  ResourcePlanIListTicketsResult,
  ResourcePlanTicketByIdResult,
  IPlanTicket,
} from '@/typings/resourcePlan';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useResourcePlanStore = defineStore({
  id: 'resourcePlanStore',
  state: () => ({}),
  actions: {
    getDiskTypes() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/disk_type/list`);
    },
    // 查询OBS项目类型列表。
    getObsProjects() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/obs_project/list`);
    },
    // 查询资源预测单据。
    reqListTickets(data: IListTicketsParam): Promise<ResourcePlanIListTicketsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/resource/ticket/list`, data);
    },
    // 获取资源预测申请单据详情。
    getTicketById(id: string): Promise<ResourcePlanTicketByIdResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/resource/ticket/${id}`);
    },
    createPlan(data: IPlanTicket) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/resource/ticket/create`, data);
    },
    getBizOrgRelation(bizId: number) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bizId}/org/relation`);
    },
    getDemandClasses() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/demand_class/list`);
    },
    getProjectTypes() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/obs_project/list`);
    },
    getRegions() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/region/list`);
    },
    getZones() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/zone/list`);
    },
    getSources() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/demand_source/list`);
    },
    getDeviceClasses() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/device_class/list`);
    },
    getDeviceTypes(device_classes: string[]) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/device_type/list`, { device_classes });
    },
  },
});
