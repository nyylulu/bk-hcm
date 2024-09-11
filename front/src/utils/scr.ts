const capacityLevel = (capacityFlag: string | number) => {
  if (capacityFlag === 4) {
    return {
      class: 'c-success',
      text: '库存充足 (50 以上)',
    };
  }

  if (capacityFlag === 3) {
    return {
      class: 'c-warning',
      text: '少量库存 (11~50)',
    };
  }

  if (capacityFlag === 2) {
    return {
      class: 'c-danger',
      text: '库存紧张 (1~10)',
    };
  }

  if (capacityFlag === 0) {
    return {
      class: 'c-disabled',
      text: '无库存',
    };
  }

  return {
    class: '',
    text: '-',
  };
};
export { capacityLevel };

function camelToSnake(str: string): string {
  return str.replace(/([A-Z])/g, '_$1').toLowerCase();
}

/**
 * 将对象key值由小驼峰命名变成下划线命名
 * @param obj
 * @returns
 */
export function convertKeysToSnakeCase(obj: any): any {
  if (Array.isArray(obj)) {
    return obj.map(convertKeysToSnakeCase);
  }
  if (obj !== null && obj?.constructor === Object) {
    return Object.keys(obj).reduce((acc: any, key: string) => {
      const snakeCaseKey = camelToSnake(key);
      acc[snakeCaseKey] = convertKeysToSnakeCase(obj[key]);
      return acc;
    }, {});
  }
  return obj;
}
