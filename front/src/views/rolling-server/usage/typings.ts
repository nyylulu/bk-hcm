import { RollingServerRecordItem } from '@/store';
import { PaginationType } from '@/typings';

export enum IView {
  ORDER = 'order',
  BIZ = 'biz',
}

export interface ISearchCondition {
  created_at?: string;
  bk_biz_id?: number;
  order_id?: string;
  [key: string]: any;
}

export interface ISearchProps {
  view: IView;
  condition: ISearchCondition;
}

export interface IDataListProps {
  view: IView;
  list: RollingServerRecordItem[];
  pagination: PaginationType;
}
