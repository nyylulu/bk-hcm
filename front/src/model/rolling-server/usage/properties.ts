import i18n from '@/language/i18n';
import { ModelProperty } from '@/model/typings';

const { t } = i18n.global;

export const appliedProperties = [
  { id: 'created_at', name: t('单据创建日期'), type: 'datetime', index: 1 },
  { id: 'bk_biz_id', name: t('业务'), type: 'bizs', index: 1 },
  { id: 'order_id', name: t('单据ID'), type: 'array', index: 1 },
] as ModelProperty[];
