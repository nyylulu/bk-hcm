import http from '@/http';
import { getEntirePath } from '@/utils';

export const updateSubnetProperties = ({ ids, properties }, config) => {
  return http.put(getEntirePath('config/updatemany/config/cvm/subnet/property'), { ids, properties }, config);
};

/**
 * 获取地域列表
 * @param {String} vendor 供应方，idc|qcloud
 * @returns {Promise}
 */
export const getRegions = async (vendor, config) => {
  const { data } = await http.get(getEntirePath(`config/find/config/${vendor}/region`), {
    ...config,
  });
  return data;
};

/**
 * 获取可用区
 * @param {String} vendor 供应方，idc|qcloud
 * @param {Array} region
 * @returns {Promise}
 */
export const getZones = ({ vendor, params }, config) => {
  return http.post(getEntirePath(`config/findmany/config/${vendor}/zone`), params, {
    removeEmptyFields: true,
    ...config,
  });
};
