import conditionCommon, { type FactoryType } from './condition-common';
import conditionApplied from './condition-applied';
import { IView } from '../../typings';

export default function optionFactory(view?: Extract<IView, IView.APPLIED>): FactoryType {
  const optionMap = {
    [IView.APPLIED]: conditionApplied,
  };
  return optionMap[view] ?? conditionCommon;
}
