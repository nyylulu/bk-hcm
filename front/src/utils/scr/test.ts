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
  if (obj !== null && obj.constructor === Object) {
    return Object.keys(obj).reduce((acc: any, key: string) => {
      const snakeCaseKey = camelToSnake(key);
      acc[snakeCaseKey] = convertKeysToSnakeCase(obj[key]);
      return acc;
    }, {});
  }
  return obj;
}
