import http from '@/http';
import { getEntirePath } from '@/utils';
// 回收单据 - 状态列表
export const getRecycleStatusOpts = () => {
  return http.get(getEntirePath('task/find/config/recycle/status'));
};

export const getRecycleStageOpts = async () => {
  const { data } = await http.get(getEntirePath('task/find/config/recycle/stage'));
  return data;
};

export const getRecycleOrders = async (params, config) => {
  const { data } = await http.post(getEntirePath('task/findmany/recycle/order'), params, config);
  return data;
};

/** 资源回收预检任务重试接口 */
export const retryOrder = ({ suborderId }: any, config: any) => {
  return http.post(
    getEntirePath('task/start/recycle/detect'),
    {
      suborder_id: suborderId,
    },
    config,
  );
};

/** 资源回收去除预检失败IP提交接口 */
export const submitOrder = ({ suborderId }, config) => {
  return http.post(
    getEntirePath('task/revise/recycle/order'),
    {
      suborder_id: suborderId,
    },
    config,
  );
};

/** 资源回收单据终止接口 */
export const stopOrder = ({ suborderId }, config) => {
  return http.post(
    getEntirePath('task/terminate/recycle/order'),
    {
      suborder_id: suborderId,
    },
    config,
  );
};

/**
 * 获取回收单据中的主机
 * @returns {Promise}
 */
export const getRecycleHosts = (params, config) => {
  return http.post(getEntirePath('task/findmany/recycle/host'), params, config);
};

/** 已回收设备机型列表查询接口 */
export const getDeviceTypeList = async () => {
  const { data } = await http.get(getEntirePath('task/find/recycle/record/devicetype'));
  return data;
};

/** 已回收设备地域列表查询接口 */
export const getRegionList = async () => {
  const { data } = await http.get(getEntirePath('task/find/recycle/record/region'));
  return data;
};

/** 已回收设备园区列表查询接口 */
export const getZoneList = async () => {
  const { data } = await http.get(getEntirePath('task/find/recycle/record/zone'));
  return data;
};

/** 资源回收单据审批接口 */
export const auditOrder = ({ suborderId, approval, remark }, config) => {
  return http.post(
    getEntirePath('task/audit/recycle/order'),
    {
      suborder_id: suborderId,
      approval,
      remark,
    },
    config,
  );
};
