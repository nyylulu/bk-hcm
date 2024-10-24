import { PropertyColumnConfig, ModelPropertyColumn } from '@/model/typings';
import { appliedProperties } from '@/model/rolling-server/usage/properties';
import { IView } from '../../typings';
import { appliedFieldIds } from './column-applied';

const columnIds = new Map<IView, string[]>();

// TODO: 可以与baseFieldIds合并到一起
const columnConfig: Record<string, PropertyColumnConfig> = {
  created_at: {
    sort: true,
  },
  applied_core: {
    sort: true,
  },
  delivered_core: {
    sort: true,
  },
};

columnIds.set(IView.APPLIED, appliedFieldIds);

// TODO: 可以放到model中定义一个view
const usageViewProperties: ModelPropertyColumn[] = [...appliedProperties];

export const getColumnIds = (resourceType: IView) => {
  return columnIds.get(resourceType);
};

const getColumns = (type: IView) => {
  const columnIds = getColumnIds(type);
  return columnIds.map((id) => ({
    ...usageViewProperties.find((item) => item.id === id),
    ...columnConfig[id],
  }));
};

const factory = {
  getColumns,
};

export type FactoryType = typeof factory;

export default factory;
