import http from '@/http';
import { getEntirePath } from '@/utils';
/**
 * 获取需求类型列表
 * @returns {Promise}
 */
export const getRequireTypes = () => {
  return http.get(getEntirePath('config/find/config/requirement'));
};
