import { has } from 'lodash';

export const useFieldVal = () => {
  let fieldList = new Map();
  const requireFieldModules = require.context('./modules', false, /\w+\.(ts)$/);

  requireFieldModules.keys().forEach((fileName: string) => {
    const fieldModule = requireFieldModules(fileName);
    const group = fileName.match(/\w+\.(ts)$/)[0].replace('.ts', '');
    const fieldModuleMap = Object.keys(fieldModule).reduce((acc, key) => {
      const field = fieldModule[key];
      return (acc = [...acc, [key, { group, ...field }]]);
    }, []);
    fieldList = new Map([...fieldList, ...fieldModuleMap]);
  });

  const convertToCamelCase = (str: string) => {
    if (!str.includes('_')) {
      return str;
    }
    return str.toLowerCase().replace(/_(.)/g, (_match: any, group1: any) => {
      return group1.toUpperCase();
    });
  };

  const getFieldCn = (fieldKey: string) => {
    return fieldList.get(convertToCamelCase(fieldKey))?.cn || fieldKey;
  };

  const getFieldCnVal = (fieldKey: string, fieldValue: any, row?: any) => {
    const field = fieldList.get(convertToCamelCase(fieldKey));
    let formattedField = fieldValue;

    if (has(field, 'transformer')) {
      return field.transformer(fieldValue, row) || '--';
    }

    if (field?.suffix) {
      formattedField = `${fieldValue}${field.suffix}`;
    }

    if (typeof fieldValue === 'string') return formattedField || '--';

    return formattedField || '--';
  };

  return { getFieldCn, getFieldCnVal };
};
