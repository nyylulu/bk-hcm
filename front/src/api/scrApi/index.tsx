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
 * 获取可用区
 * @param {String} vendor 供应方，idc|qcloud
 * @param {Array} region
 * @returns {Promise}
 */
const getZones = async ({ vendor, region = [], isCmdbRegion = false }, config) => {
  const params = {};

  if (vendor === 'qcloud') {
    if (isCmdbRegion) {
      params.cmdb_region_name = region;
    } else {
      params.region = region.length ? region : undefined;
    }
  } else if (vendor === 'idc') {
    params.cmdb_region_name = region;
  }

  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/${vendor}/zone`,
    params,
    {
      removeEmptyFields: true,
      ...config,
    },
  );
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
const getImages = async (params) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/image`, { params });
  return data;
};
/**
 * 获取数据盘类型列表
 */
const getDiskTypes = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/cvm/disktype`);
  return data;
};
const getAntiAffinityLevels = async (resourceType: any, hasZone: any, config: any) => {
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/affinity`,
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
const getRecyclableHosts = async ({ bk_biz_id, ips, asset_ids, bk_host_ids }: any, config: any) => {
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/findmany/recycle/recyclability`,
    {
      bk_biz_id,
      ips,
      asset_ids,
      bk_host_ids,
    },
    config,
  );
  return data;
};
/** 业务待回收主机列表查询接口 */
const getRecycleList = async ({ bkBizId, page }, config) => {
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/task/find/recycle/biz/host`,
    {
      bk_biz_id: bkBizId,
      page,
    },
    config,
  );
  return data;
};
/**
 * 搜索栏业务列表
 */
const getBusinesses = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/auth/apps`);
  return data;
};
/**
 * 搜索栏业务列表新增业务跳转URL
 */
const getAuthApplyUrl = async (permission) => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/auth/apply`, {
    ...permission,
  });
  return data;
};
/**
 * 搜索栏需求类型列表
 */
const getRequireTypes = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/requirement`);
  return data;
};
/**
 * 获取地域列表
 * @param {String} vendor 供应方，idc|qcloud
 * @returns {Promise}
 */
const getRegions = async (vendor: string): Promise<any> => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/${vendor}/region`);
  return data;
};
/**
 * 获取设备类型
 * @returns {Promise}
 */
const getDeviceTypes = async ({ region, zone, require_type, device_group, enable_capacity }) => {
  const rules = [
    region.length && { field: 'region', operator: 'in', value: region },
    zone.length && { field: 'zone', operator: 'in', value: zone },
    require_type && { field: 'require_type', operator: 'equal', value: require_type },
    device_group && { field: 'label.device_group', operator: 'in', value: device_group },
    enable_capacity && { field: 'enable_capacity', operator: 'equal', value: enable_capacity },
  ].filter(Boolean);
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/devicetype`,
    {
      filter: {
        condition: 'AND',
        rules,
      },
    },
    {
      simpleConditions: true,
      removeEmptyFields: true,
    },
  );
  return data;
};

/**
 * 获取 cpu mem disk 可选项
 * @returns {Promise}
 */
const getRestrict = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/cvm/devicerestrict`);
  return data;
};
/**
 * 获取物理机机型
 */
const getIDCPMDeviceTypes = async () => {
  const { data } = await http.post(
    `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/idcpm/devicetype`,
    { page: {} },
    { simpleConditions: true },
  );
  return data;
};
/**
 * 获取物理机操作系统
 */
const getOsTypes = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/idcpm/ostype`);
  return data;
};
/**
 * 获取物理机运营商
 */
const getIsps = async () => {
  const { data } = await http.get(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/find/config/idcpm/isp`);
  return data;
};
// VPC列表
const getVpcs = async (region: any) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/vpc`, {
    region,
  });
  return data;
};
// 子网列表
const getSubnets = async ({ region, zone, vpc }) => {
  const { data } = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/config/findmany/config/cvm/subnet`, {
    region,
    zone,
    vpc,
  });
  return data;
};
export default {
  getAreas,
  getZones,
  getCvmTypes,
  getImages,
  getDiskTypes,
  getAntiAffinityLevels,
  getRecyclableHosts,
  getRecycleList,
  getBusinesses,
  getAuthApplyUrl,
  getRequireTypes,
  getRegions,
  getDeviceTypes,
  getVpcs,
  getSubnets,
  getIDCPMDeviceTypes,
  getOsTypes,
  getIsps,
  getRestrict,
};
