import http from '@/http';
import { CreateRecallTaskModal } from '@/typings/scr';
import { defineStore } from 'pinia';

import type { IPageQuery } from '@/typings/common';
import type {
  IRecycleArea,
  IQueryDissolveList,
  IDissolveList,
  IDissolveHostCurrentListResult,
  IDissolveHostCurrentListParam,
  IDissolveHostOriginListResult,
  IDissolveHostOriginListParam,
} from '@/typings/ziyanScr';
import { transferSimpleConditions } from '@/utils/scr/simple-query-builder';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export const useZiyanScrStore = defineStore('ziyanScr', () => {
  const listVpc = (region: any) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/mov/cvm/manage/describevpcs`, { region });
  };
  const listSubnet = ({ region, zone, vpcId }) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/mov/cvm/manage/describesubnets`, { region, zone, vpcId });
  };

  /**
   * @returns 资源上下架 - 获取单据状态list
   */
  const getTaskStatusList = () => {
    return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/pool/find/config/task/status`);
  };

  /**
   * @returns 机型配置信息列表
   */
  const getDeviceTypeList = () => {
    return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/pool/find/config/devicetype`);
  };

  /**
   * @returns IDC物理机操作系统列表
   */
  const getIdcpmOsTypeList = () => {
    return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/idcpm/ostype`);
  };

  /**
   * @returns IDC地域列表
   */
  const getIdcRegionList = () => {
    return http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/idc/region`);
  };

  /**
   * @param data.cmdb_region_name 地域列表。若列表非空，则返回地域列表下的区域信息；若列表为空，则返回所有地域下的区域信息
   * @returns IDC可用区配置信息列表
   */
  const queryIdcZoneList = (data: { cmdb_region_name: string[] }) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/idc/zone`, data);
  };

  const getRecycleAreas = (page: IPageQuery): Promise<{ data: { detail: IRecycleArea[] } }> => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/dissolve/recycled_module/list`, { page });
  };

  const getDissolveList = (data: IQueryDissolveList): Promise<IDissolveList> => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/dissolve/table/list`, data);
  };

  /**
   * 资源申请单据执行接口
   * @returns {Promise}
   */
  const retryOrder = (data: any) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/start/apply`, data);
  };

  /**
   * 资源申请单据终止接口
   * @returns {Promise}
   */
  const stopOrder = (data: any) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/terminate/apply`, data);
  };

  /**
   * 资源从资源池下架
   */
  const createRecallTask = (data: CreateRecallTaskModal) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/pool/create/recall/task`, data);
  };

  /**
   * 资源上架到资源池
   * @param data 要上架的CC主机ID，数量最大500
   */
  const createOnlineTask = (data: { bk_host_ids: string[] }) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/pool/create/launch/task`, data);
  };

  // 资源生产详情
  const getProductionDetails = (subOrderId: any, page: any, status: any) =>
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/find/apply/record/generate`, {
      suborder_id: subOrderId,
      page,
      filter: status || status === 0 ? transferSimpleConditions(['AND', ['status', '=', status]]) : undefined,
    });

  // 资源初始化详情
  const getInitializationDetails = (subOrderId: any, page: any, keyword: any, status: any) =>
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/find/apply/record/init`, {
      suborder_id: subOrderId,
      page,
      filter:
        status || status === 0
          ? transferSimpleConditions(['AND', ['ip', 'contains', keyword], ['status', '=', status]])
          : undefined,
    });

  // 本地盘性能压测
  const getDiskCheckDetails = (subOrderId: any, page: any, keyword: any, status: any) =>
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/find/apply/record/disk_check`, {
      suborder_id: subOrderId,
      page,
      filter:
        status || status === 0
          ? transferSimpleConditions(['AND', ['ip', 'contains', keyword], ['status', '=', status]])
          : undefined,
    });

  // 资源交付详情
  const getDeliveryDetails = (subOrderId: any, page: any, keyword: any, status: any) =>
    http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/find/apply/record/deliver`, {
      suborder_id: subOrderId,
      page,
      filter:
        status || status === 0
          ? transferSimpleConditions(['AND', ['ip', 'contains', keyword], ['status', '=', status]])
          : undefined,
    });

  // 查询裁撤数据中当前主机信息。
  const dissolveHostCurrentList = (data: IDissolveHostCurrentListParam): Promise<IDissolveHostCurrentListResult> => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/dissolve/host/current/list`, data);
  };

  // 查询裁撤数据中原始主机信息。
  const dissolveHostOriginList = (data: IDissolveHostOriginListParam): Promise<IDissolveHostOriginListResult> => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/dissolve/host/origin/list`, data);
  };
  return {
    listVpc,
    listSubnet,
    getTaskStatusList,
    getDeviceTypeList,
    getIdcpmOsTypeList,
    getIdcRegionList,
    queryIdcZoneList,
    getRecycleAreas,
    getDissolveList,
    retryOrder,
    stopOrder,
    createRecallTask,
    createOnlineTask,
    getProductionDetails,
    getInitializationDetails,
    getDiskCheckDetails,
    getDeliveryDetails,
    dissolveHostCurrentList,
    dissolveHostOriginList,
  };
});
