import http from '@/http';
import { defineStore } from 'pinia';

import type { IPageQuery } from '@/typings/common';
import type { IRecycleArea } from '@/typings/ziyanScr';

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

  return {
    listVpc,
    listSubnet,
    getTaskStatusList,
    getDeviceTypeList,
    getIdcpmOsTypeList,
    getIdcRegionList,
    queryIdcZoneList,
    getRecycleAreas,
  };
});
