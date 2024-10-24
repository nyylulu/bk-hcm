import { appliedProperties } from '@/model/rolling-server/usage/properties';
import { appliedFieldIds } from './condition-applied';
import { IView } from '../../typings';

const conditionFieldIds = new Map<IView, string[]>();
conditionFieldIds.set(IView.APPLIED, appliedFieldIds);

const usageViewProperties = [...appliedProperties];

export const getConditionFieldIds = (view: IView) => {
  return conditionFieldIds.get(view);
};

const getConditionField = (view: IView) => {
  const fieldIds = getConditionFieldIds(view);
  const fields = fieldIds.map((id) => usageViewProperties.find((item) => item.id === id));
  return fields;
};

const factory = {
  getConditionField,
};

export type FactoryType = typeof factory;

export default factory;
