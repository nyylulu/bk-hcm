import http from '@/http';

/**
 * 获取需求类型列表
 * @returns {Promise}
 */
export const getRequireTypes = () => {
  return http.get('config/find/config/requirement');
};
