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
  IListResourcesDemandsParam,
  IListResourcesDemandsResult,
  IPlanDemandResult,
  IListChangeLogsParam,
  IListChangeLogsResult,
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
    // 获取预测类型列表
    getDemandClasses() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/demand_class/list`);
    },
    // 获取项目类型列表
    getProjectTypes() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/obs_project/list`);
    },
    // 获取城市列表
    getRegions() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/region/list`);
    },
    // 获取可用区列表
    getZones(region_ids?: string[]) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/zone/list`, { region_ids });
    },
    getSources() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/demand_source/list`);
    },
    // 获取机型规格列表
    getDeviceClasses() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/device_class/list`);
    },
    // 获取机型类型列表
    getDeviceTypes(device_classes?: string[]) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/meta/device_type/list`, { device_classes });
    },
    // 获取运营产品列表
    getOpProducts() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/op_products/list`);
    },
    // 获取规划产品列表
    getPlanProducts() {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/plan_products/list`);
    },
    // 获取计划类型列表
    getPlanTypes() {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/metas/plan_types/list`);
    },
    // 查询资源预测单据。
    reqListTickets(data: IListTicketsParam): Promise<ResourcePlanIListTicketsResult> {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/resource/ticket/list`, data);
    },
    // 获取资源预测申请单据详情。
    getTicketById(id: string): Promise<ResourcePlanTicketByIdResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/resource/ticket/${id}`);
    },
    // 查询资源预测单据状态列表。
    getStatusList(): Promise<IResPlanTicketStatusListResult> {
      return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plan/res_plan_ticket_status/list`);
    },
    // 查询业务下资源预测需求列表
    getResourcesDemandsList(bk_biz_id: number, data: IListResourcesDemandsParam): Promise<IListResourcesDemandsResult> {
      http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bk_biz_id}/plans/resources/demands/list`, data);
      return new Promise((resolve) =>
        setTimeout(() => {
          resolve({
            // @ts-ignore
            data: {
              count: 2,
              overview: {
                total_cpu_core: 1024,
                total_applied_core: 1024,
                in_plan_cpu_core: 512,
                in_plan_applied_cpu_core: 512,
                out_plan_cpu_core: 512,
                out_plan_applied_cpu_core: 512,
                expiring_cpu_core: 224,
              },
              details: [
                {
                  crp_demand_id: 387330,
                  bk_biz_id: 111,
                  bk_biz_name: '业务',
                  op_product_id: 222,
                  op_product_name: '运营产品',
                  status: 'can_apply',
                  status_name: '变更中',
                  demand_class: 'CVM',
                  available_year_month: '2024-01',
                  expect_time: '2024-01-01',
                  device_class: '高IO型I6t',
                  device_type: 'I6t.33XMEDIUM198',
                  total_os: 56,
                  applied_os: 44,
                  remained_os: 12,
                  total_cpu_core: 560,
                  applied_cpu_core: 440,
                  remained_cpu_core: 120,
                  total_memory: 560,
                  applied_memory: 440,
                  remained_memory: 120,
                  total_disk_size: 560,
                  applied_disk_size: 440,
                  remained_disk_size: 120,
                  region_id: 'ap-shanghai',
                  region_name: '上海',
                  zone_id: 'ap-shanghai-2',
                  zone_name: '上海二区',
                  plan_type: '预测内',
                  obs_project: '常规项目',
                  generation_type: '采购',
                  device_family: '高IO型',
                  disk_type: 'CLOUD_PREMIUM',
                  disk_type_name: '高性能云硬盘',
                  disk_io: 15,
                },
                {
                  crp_demand_id: 387330,
                  bk_biz_id: 111,
                  bk_biz_name: '业务',
                  op_product_id: 222,
                  op_product_name: '运营产品',
                  status: 'can_apply',
                  status_name: '变更中',
                  demand_class: 'CVM',
                  available_year_month: '2024-01',
                  expect_time: '2024-01-01',
                  device_class: '高IO型I6t',
                  device_type: 'I6t.33XMEDIUM198',
                  total_os: 56,
                  applied_os: 44,
                  remained_os: 12,
                  total_cpu_core: 560,
                  applied_cpu_core: 440,
                  remained_cpu_core: 120,
                  total_memory: 560,
                  applied_memory: 440,
                  remained_memory: 120,
                  total_disk_size: 560,
                  applied_disk_size: 440,
                  remained_disk_size: 120,
                  region_id: 'ap-shanghai',
                  region_name: '上海',
                  zone_id: 'ap-shanghai-2',
                  zone_name: '上海二区',
                  plan_type: '预测内',
                  obs_project: '常规项目',
                  generation_type: '采购',
                  device_family: '高IO型',
                  disk_type: 'CLOUD_PREMIUM',
                  disk_type_name: '高性能云硬盘',
                  disk_io: 15,
                },
              ],
            },
          });
        }, 2000),
      );
    },
    // 查询业务下资源预测需求详情信息
    getPlanDemand(bk_biz_id: number, crp_demand_id: number): Promise<IPlanDemandResult> {
      http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bk_biz_id}/plans/demands/${crp_demand_id}`);
      return new Promise((resolve) =>
        setTimeout(() => {
          resolve({
            crp_demand_id: '387330',
            year_month_week: '2024年10月4周',
            expect_start_date: '2024-10-21',
            expect_end_date: '2024-10-27',
            expect_time: '2024-10-21',
            bk_biz_id: 111,
            bk_biz_name: '业务',
            bg_id: 4,
            bg_name: 'IEG互动娱乐事业群',
            dept_id: 1041,
            dept_name: 'IEG技术运营部',
            plan_product_id: 34,
            plan_product_name: '规划产品',
            op_product_id: 41,
            op_product_name: '运营产品',
            obs_project: '常规项目',
            area_id: 'south',
            area_name: '华南地区',
            region_id: 'guangzhou',
            region_name: '广州',
            zone_id: 'guangzhou-3',
            zone_name: '广州三区',
            plan_type: '计划内',
            plan_advance_week: 9,
            expedited_postponed: '无变化',
            core_type_id: 1,
            core_type: '小核心',
            device_family: '标准型',
            device_class: '标准型S5',
            device_type: 'S5.2XLARGE16',
            os: 0.125,
            memory: 2.0,
            cpu_core: 1,
            disk_size: 1,
            disk_io: 150,
            disk_type: 'CLOUD_PREMIUM',
            disk_type_name: '高性能云硬盘',
            demand_week: 'UNPLAN_9_13W',
            res_pool_type: 0,
            res_pool: '自研池',
            res_mode: '按机型',
            generation_type: '采购',
          });
        }, 2000),
      );
    },
    // 查询资源预测需求单的变更历史
    getListChangeLogs(bk_biz_id: number, data: IListChangeLogsParam): Promise<IListChangeLogsResult> {
      http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bk_biz_id}/plans/demands/change_logs/list`, data);
      return new Promise((resolve) =>
        setTimeout(() => {
          resolve({
            // @ts-ignore
            data: {
              count: 2,
              details: [
                {
                  crp_demand_id: 387330,
                  expect_time: '2024-10-21',
                  bg_name: 'IEG互动娱乐事业群',
                  dept_name: 'IEG技术运营部',
                  plan_product_name: '移动终端游戏',
                  op_product_name: '运营产品',
                  obs_project: '常规项目',
                  region_name: '广州',
                  zone_name: '广州三区',
                  demand_week: 'UNPLAN_9_13W',
                  res_pool_type: 0,
                  device_class: '标准型S5',
                  device_type: 'S5.2XLARGE16',
                  change_cvm_amount: 0.125,
                  after_cvm_amount: 0.125,
                  change_core_amount: 1,
                  after_core_amount: 1,
                  change_ram_amount: 2,
                  after_ram_amount: 2,
                  disk_type: null,
                  disk_io: 0,
                  changed_disk_amount: 1,
                  after_disk_amount: 1,
                  demand_source: '追加需求订单',
                  crp_sn: 'XQ202408221500512986',
                  create_time: null,
                  remark: '由 UNPLAN_9_13W 自动变为 UNPLAN_9_13W\n',
                  res_pool: '自研池',
                },
              ],
            },
          });
        }, 2000),
      );
    },
    // 查询管理下资源预测需求列表
    getResourcesDemandsListByOrg(data: IListResourcesDemandsParam): Promise<IListResourcesDemandsResult> {
      http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/resources/demands/list`, data);
      return new Promise((resolve) =>
        setTimeout(() => {
          resolve({
            // @ts-ignore
            data: {
              count: 2,
              overview: {
                total_cpu_core: 1024,
                total_applied_core: 1024,
                in_plan_cpu_core: 512,
                in_plan_applied_cpu_core: 512,
                out_plan_cpu_core: 512,
                out_plan_applied_cpu_core: 512,
                expiring_cpu_core: 224,
              },
              details: [
                {
                  crp_demand_id: 387330,
                  bk_biz_id: 111,
                  bk_biz_name: '业务',
                  op_product_id: 222,
                  op_product_name: '运营产品',
                  status: 'can_apply',
                  status_name: '变更中',
                  demand_class: 'CVM',
                  available_year_month: '2024-01',
                  expect_time: '2024-01-01',
                  device_class: '高IO型I6t',
                  device_type: 'I6t.33XMEDIUM198',
                  total_os: 56,
                  applied_os: 44,
                  remained_os: 12,
                  total_cpu_core: 560,
                  applied_cpu_core: 440,
                  remained_cpu_core: 120,
                  total_memory: 560,
                  applied_memory: 440,
                  remained_memory: 120,
                  total_disk_size: 560,
                  applied_disk_size: 440,
                  remained_disk_size: 120,
                  region_id: 'ap-shanghai',
                  region_name: '上海',
                  zone_id: 'ap-shanghai-2',
                  zone_name: '上海二区',
                  plan_type: '预测内',
                  obs_project: '常规项目',
                  generation_type: '采购',
                  device_family: '高IO型',
                  disk_type: 'CLOUD_PREMIUM',
                  disk_type_name: '高性能云硬盘',
                  disk_io: 15,
                },
                {
                  crp_demand_id: 387330,
                  bk_biz_id: 111,
                  bk_biz_name: '业务',
                  op_product_id: 222,
                  op_product_name: '运营产品',
                  status: 'can_apply',
                  status_name: '变更中',
                  demand_class: 'CVM',
                  available_year_month: '2024-01',
                  expect_time: '2024-01-01',
                  device_class: '高IO型I6t',
                  device_type: 'I6t.33XMEDIUM198',
                  total_os: 56,
                  applied_os: 44,
                  remained_os: 12,
                  total_cpu_core: 560,
                  applied_cpu_core: 440,
                  remained_cpu_core: 120,
                  total_memory: 560,
                  applied_memory: 440,
                  remained_memory: 120,
                  total_disk_size: 560,
                  applied_disk_size: 440,
                  remained_disk_size: 120,
                  region_id: 'ap-shanghai',
                  region_name: '上海',
                  zone_id: 'ap-shanghai-2',
                  zone_name: '上海二区',
                  plan_type: '预测内',
                  obs_project: '常规项目',
                  generation_type: '采购',
                  device_family: '高IO型',
                  disk_type: 'CLOUD_PREMIUM',
                  disk_type_name: '高性能云硬盘',
                  disk_io: 15,
                },
              ],
            },
          });
        }, 2000),
      );
    },
    // 查询管理下资源预测需求详情信息
    getPlanDemandByOrg(crp_demand_id: number): Promise<IPlanDemandResult> {
      http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/demands/${crp_demand_id}`);
      return new Promise((resolve) =>
        setTimeout(() => {
          resolve({
            crp_demand_id: '387330',
            year_month_week: '2024年10月4周',
            expect_start_date: '2024-10-21',
            expect_end_date: '2024-10-27',
            expect_time: '2024-10-21',
            bk_biz_id: 111,
            bk_biz_name: '业务',
            bg_id: 4,
            bg_name: 'IEG互动娱乐事业群',
            dept_id: 1041,
            dept_name: 'IEG技术运营部',
            plan_product_id: 34,
            plan_product_name: '规划产品',
            op_product_id: 41,
            op_product_name: '运营产品',
            obs_project: '常规项目',
            area_id: 'south',
            area_name: '华南地区',
            region_id: 'guangzhou',
            region_name: '广州',
            zone_id: 'guangzhou-3',
            zone_name: '广州三区',
            plan_type: '计划内',
            plan_advance_week: 9,
            expedited_postponed: '无变化',
            core_type_id: 1,
            core_type: '小核心',
            device_family: '标准型',
            device_class: '标准型S5',
            device_type: 'S5.2XLARGE16',
            os: 0.125,
            memory: 2.0,
            cpu_core: 1,
            disk_size: 1,
            disk_io: 150,
            disk_type: 'CLOUD_PREMIUM',
            disk_type_name: '高性能云硬盘',
            demand_week: 'UNPLAN_9_13W',
            res_pool_type: 0,
            res_pool: '自研池',
            res_mode: '按机型',
            generation_type: '采购',
          });
        }, 2000),
      );
    },
    // 查询管理下资源预测需求单的变更历史
    getListChangeLogsByOrg(data: IListChangeLogsParam): Promise<IListChangeLogsResult> {
      http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/plans/demands/change_logs/list`, data);
      return new Promise((resolve) =>
        setTimeout(() => {
          resolve({
            // @ts-ignore
            data: {
              count: 2,
              details: [
                {
                  crp_demand_id: 387330,
                  expect_time: '2024-10-21',
                  bg_name: 'IEG互动娱乐事业群',
                  dept_name: 'IEG技术运营部',
                  plan_product_name: '移动终端游戏',
                  op_product_name: '运营产品',
                  obs_project: '常规项目',
                  region_name: '广州',
                  zone_name: '广州三区',
                  demand_week: 'UNPLAN_9_13W',
                  res_pool_type: 0,
                  device_class: '标准型S5',
                  device_type: 'S5.2XLARGE16',
                  change_cvm_amount: 0.125,
                  after_cvm_amount: 0.125,
                  change_core_amount: 1,
                  after_core_amount: 1,
                  change_ram_amount: 2,
                  after_ram_amount: 2,
                  disk_type: null,
                  disk_io: 0,
                  changed_disk_amount: 1,
                  after_disk_amount: 1,
                  demand_source: '追加需求订单',
                  crp_sn: 'XQ202408221500512986',
                  create_time: null,
                  remark: '由 UNPLAN_9_13W 自动变为 UNPLAN_9_13W\n',
                  res_pool: '自研池',
                },
              ],
            },
          });
        }, 2000),
      );
    },
    // 批量取消资源预测需求
    cancelResourcesDemands(bk_biz_id: number, data: { crp_demand_ids: number[] }) {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/bizs/${bk_biz_id}/plans/resources/demands/cancel`, data);
    // 查询单据类型列表。
    },
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
