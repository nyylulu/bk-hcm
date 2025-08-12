import http from '@/http';
import { getEntirePath } from '@/utils';
import type { cvmProduceQueryReq, maxResourceCapacity, createCvmOrder } from '@/typings/cvm-pro';

export const getCvmProduceOrderStatusOpts = () => {
  return http.get(getEntirePath('cvm/find/config/apply/status'));
};

/**
 * 获取CVM生产列表
 * @param {cvmProduceQueryReq} params 参数
 * @returns {Promise}
 */
// 注意： 不报错
export const getCvmProduceOrderList = (params: cvmProduceQueryReq) => {
  return http.post(getEntirePath('cvm/findmany/apply/order'), params);
};

/**
 * 获取 cpu mem disk 可选项
 * @returns {Promise}
 */
export const getRestrict = () => {
  return http.get(getEntirePath('config/find/config/cvm/devicerestrict'));
};

/**
 * 获取资源最大申领量
 * @param {maxResourceCapacity} params 参数
 * @returns {Promise}
 */
export const getCapacity = (params: maxResourceCapacity) => {
  return http.post(getEntirePath('config/find/cvm/capacity'), params);
};

export const getVpcs = (params) => {
  return http.post(getEntirePath('config/findmany/config/cvm/vpc'), params);
};

export const getImages = (params: { region: string[] }) => {
  return http.post(getEntirePath('config/findmany/config/cvm/image'), params);
};

export const getSubnets = ({ region, zone, vpc }) => {
  return http.post(getEntirePath('config/findmany/config/cvm/subnet'), {
    region,
    zone,
    vpc,
  });
};

/**
 * 创建CVM生产单据
 * @param {createCvmOrder} params 参数

 * @returns {Promise}
 */
export const createCvmProduceOrder = (params: createCvmOrder) => {
  return http.post(getEntirePath('cvm/create/apply/order'), params);
};

export const getDiskTypes = () => http.get(getEntirePath('config/find/config/cvm/disktype'));

// order_id
export const getCvmProducedResources = (params) => {
  return http.post(getEntirePath('cvm/findmany/apply/device'), params);
};
