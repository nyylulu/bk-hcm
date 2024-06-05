import http from '@/http';
import { defineStore } from 'pinia';

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
  return {
    listVpc,
    listSubnet,
    getTaskStatusList,
  };
});
