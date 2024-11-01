import { QuotaAdjustType } from '../typings';
import type { PaginationType } from '@/typings';
import type { IRollingServerBizQuotaItem } from '@/store/rolling-server-quota';

export interface IBizViewSearchCondition {
  quota_month?: string;
  bk_biz_ids?: number[];
  adjust_type?: QuotaAdjustType;
  reviser?: string;
  [key: string]: any;
}

export interface IBizViewSearchProps {
  condition: IBizViewSearchCondition;
}

export interface IBizViewDataListProps {
  list: IRollingServerBizQuotaItem[];
  pagination: PaginationType;
}
