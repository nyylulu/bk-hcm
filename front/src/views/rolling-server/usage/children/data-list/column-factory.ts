import columnCommon, { type FactoryType } from './column-common';
import columnApplied from './column-applied';
import { IView } from '../../typings';

export default function optionFactory(view?: Extract<IView, IView.APPLIED>): FactoryType {
  const optionMap = {
    [IView.APPLIED]: columnApplied,
  };
  return optionMap[view] ?? columnCommon;
}
