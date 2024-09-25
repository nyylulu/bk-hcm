import http from '@/http';
import { defineStore } from 'pinia';
import {
  IBizResourcesTicketsParam,
  IOpResourcesTicketsParam,
  IOpResourcesTicketsResult,
  IBizResourcesTicketsResult,
  ResourcePlanTicketByIdResult,
  IResPlanTicketStatusListResult,
  IPlanTicket,
  ITicketTypesResult,
  IOpProductsResult,
  IPlanProductsResult,
  IBizsByOpProductResult,
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
    // 业务视角查询资源预测单据。
    getBizResourcesTicketsList(bizId: number, data: IBizResourcesTicketsParam): Promise<IBizResourcesTicketsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bizId}/plans/resources/tickets/list`, data);
    },
    // 管理员查询资源预测单据。
    getOpResourcesTicketsList(data: IOpResourcesTicketsParam): Promise<IOpResourcesTicketsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/resources/tickets/list`, data);
    },
    // 获取 资源管理 资源预测申请单据详情。
    getBizResourcesTicketsById(bizId: number, id: string): Promise<ResourcePlanTicketByIdResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bizId}/plans/resources/tickets/${id}`);
    },
    // 获取  服务请求 资源预测申请单据详情。
    getOpResourcesTicketsById(id: string): Promise<ResourcePlanTicketByIdResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/resources/tickets/${id}`);
    },
    createPlan(data: IPlanTicket): { data: { id: string } } {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/resource/ticket/create`, data);
    },
    getBizOrgRelation(bizId: number) {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bizId}/org/relation`);
    },
    getDemandClasses() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/demand_class/list`);
    },
    getRegions() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/region/list`);
    },
    getZones(region_ids: string[]) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/zone/list`, { region_ids });
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
    // 查询资源预测单据状态列表。
    getStatusList(): Promise<IResPlanTicketStatusListResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/res_plan_ticket_status/list`);
    },
    // 查询单据类型列表。
    getTicketTypesList(): Promise<ITicketTypesResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/ticket_types/list`);
    },
    // 查询运营产品列表。
    getOpProductsList(): Promise<IOpProductsResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/op_products/list`);
    },
    // 查询规划产品列表。
    getPlanProductsList(): Promise<IPlanProductsResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/plan_products/list`);
    },
    // 根据运营产品ID查询业务列表。
    getBizsByOpProductList(data: { op_product_id: number }): Promise<IBizsByOpProductResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/bizs/by/op_product/list`, data);
    },
  },
});
