import { get as _get } from 'lodash-es';

export const lodashGet = (obj: any, path: string, defaultValue?: any) => {
  if (typeof obj !== 'object' || !path) {
    return obj;
  }
  return _get(obj, path, defaultValue);
};
