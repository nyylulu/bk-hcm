import type { PaginationType } from '@/typings';
import type { IRollingServerBillItem } from '@/store/rolling-server-bills';

export interface IBillsSearchCondition {
  date?: string;
  bk_biz_id?: number;
  [key: string]: any;
}

export interface IBillsSearchProps {
  condition: IBillsSearchCondition;
}

export interface IBillsDataListProps {
  list: IRollingServerBillItem[];
  pagination: PaginationType;
}
