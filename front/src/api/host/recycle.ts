import http from '@/http';

// 回收单据 - 状态列表
export const getRecycleStatusOpts = () => {
  return http.get('task/find/config/recycle/status');
};

export const getRecycleStageOpts = () => {
  return http.get('task/find/config/recycle/stage');
};

export const getRecycleOrders = (params, config) => {
  return http.post('task/findmany/recycle/order', params, {
    removeEmptyFields: true,
    transformFields: true,
    ...config,
  });
};

/** 资源回收预检任务重试接口 */
export const retryOrder = ({ suborderId }: any, config: any) => {
  return http.post(
    'task/start/recycle/order',
    {
      suborder_id: suborderId,
    },
    config,
  );
};

/** 资源回收去除预检失败IP提交接口 */
export const submitOrder = ({ suborderId }, config) => {
  return http.post(
    'task/revise/recycle/order',
    {
      suborder_id: suborderId,
    },
    config,
  );
};

/** 资源回收单据终止接口 */
export const stopOrder = ({ suborderId }, config) => {
  return http.post(
    'task/terminate/recycle/order',
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
  return http.post('task/findmany/recycle/host', params, {
    transformFields: true,
    removeEmptyFields: true,
    ...config,
  });
};

/** 已回收设备机型列表查询接口 */
export const getDeviceTypeList = () => {
  return http.get('task/find/recycle/record/devicetype');
};

/** 已回收设备地域列表查询接口 */
export const getRegionList = () => {
  return http.get('task/find/recycle/record/region');
};

/** 已回收设备园区列表查询接口 */
export const getZoneList = () => {
  return http.get('task/find/recycle/record/zone');
};
