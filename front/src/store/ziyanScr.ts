import http from '@/http';
import { defineStore } from 'pinia';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;
export const useziyanScrStore = defineStore('ziyanScr', () => {
  const listVpc = (region: any) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/mov/cvm/manage/describevpcs`, { region });
  };
  const listSubnet = ({ region, zone, vpcId }) => {
    return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/mov/cvm/manage/describesubnets`, { region, zone, vpcId });
  };
  return {
    listVpc,
    listSubnet,
  };
});
