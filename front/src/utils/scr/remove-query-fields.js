import { isArray, isObject, isEmpty } from 'lodash';

export const removeEmptyFields = (data) => {
  Object.keys(data).forEach((key) => {
    const value = data[key];

    if (typeof value === 'number' || typeof value === 'boolean') return;

    if (isEmpty(value)) {
      delete data[key];
      return;
    }

    if (isArray(value)) {
      value.forEach((item) => {
        removeEmptyFields(item);
      });
      return;
    }

    if (isObject(value)) {
      const result = removeEmptyFields(value);

      if (isEmpty(result)) {
        delete data[key];
      }
    }
  });

  return data;
};
