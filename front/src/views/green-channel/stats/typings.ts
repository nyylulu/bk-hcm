import type { PaginationType } from '@/typings';
import type { IRollingServerBizQuotaItem } from '@/store/rolling-server-quota';

export interface ISearchCondition {
  bk_biz_ids?: number[];
  [key: string]: any;
}

export interface ISearchProps {
  condition: ISearchCondition;
}

export interface IDataListProps {
  list: IRollingServerBizQuotaItem[];
  pagination: PaginationType;
}
