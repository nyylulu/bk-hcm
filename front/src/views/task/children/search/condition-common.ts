import { ResourceTypeEnum } from '@/common/resource-constant';
import { getModel } from '@/model/manager';
import { SearchClbView, SearchCvmView } from '@/model/task/search.view';

const getConditionField = (type: ResourceTypeEnum) => {
  if (type === ResourceTypeEnum.CLB) {
    return getClbConditionField();
  }

  if (type === ResourceTypeEnum.CVM) {
    return getCvmConditionField();
  }
};

const getClbConditionField = () => {
  const properties = getModel(SearchClbView).getProperties();
  return properties.filter((item) => item.id !== 'resource');
};

const getCvmConditionField = () => {
  const properties = getModel(SearchCvmView).getProperties();
  return properties.filter((item) => item.id !== 'resource');
};

const factory = {
  getConditionField,
};

export type FactoryType = typeof factory;

export default factory;
