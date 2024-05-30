import http from '@/http';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

/**
 * 获取区域列表
 */
const getAreas = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/cvm/tablelist/arealist`);
  return data;
};
/**
 * 获取可用区列表
 * @param  {String} area 区域
 */
const getZones = async (area: string) => {
  const params = {
    area,
  };
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/cvm/tablelist/zonelist`, { params });
  return data;
};
/**
 * 获取 CVM 机型列表
 * @param  {String} zone 可用区 id
 */
const getCvmTypes = async (zone: string, area: any) => {
  const params = {
    zone,
    area,
  };
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/cvm/tablelist/cvmtype`, { params });
  return data;
};
/**
 * 获取镜像列表
 * @param  {String} area 区域 id
 */
const getImages = async (area: string) => {
  const params = {
    area,
  };
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/cvm/tablelist/imagename`, { params });
  return data;
};
/**
 * 获取数据盘类型列表
 */
const getDiskTypes = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/cvm/tablelist/datadisktype`);
  return data;
};
const getAntiAffinityLevels = async (resourceType, hasZone, config) => {
  const { data } = await http.get(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/mov/config/find/config/affinity`,
    {
      resource_type: resourceType,
      has_zone: hasZone,
    },
    {
      config,
    },
  );
  return data;
};

export default { getAreas, getZones, getCvmTypes, getImages, getDiskTypes, getAntiAffinityLevels };
