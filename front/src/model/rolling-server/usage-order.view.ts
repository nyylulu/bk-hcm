import { APPLIED_TYPE_NAME, RETURNED_WAY_NAME } from '@/views/rolling-server/usage/constants';
import { QueryRuleOPEnum } from '@/typings';
import { ModelProperty } from '../typings';
import i18n from '@/language/i18n';
import dayjs from 'dayjs';

const { t } = i18n.global;

export default [
  { id: 'id', name: t('ID'), type: 'string' },
  { id: 'bk_biz_id', name: t('业务'), type: 'business' },
  { id: 'order_id', name: t('单据ID'), type: 'string' },
  { id: 'suborder_id', name: t('子单据ID'), type: 'string', meta: { search: { op: QueryRuleOPEnum.IN } } },
  { id: 'year', name: t('申请时间年份'), type: 'string' },
  { id: 'month', name: t('申请时间月份'), type: 'string' },
  { id: 'day', name: t('申请时间天'), type: 'string' },
  { id: 'creator', name: t('创建者'), type: 'user' },
  {
    id: 'roll_date',
    name: t('单据日期'),
    type: 'datetime',
    meta: {
      search: {
        format(value: Date[] | Date | string) {
          // TODO: 数组时不转换，当qs获取时传入的是数组
          if (Array.isArray(value)) {
            return value;
          }
          const date = dayjs(value);
          return Number(date.format('YYYYMMDD'));
        },
      },
    },
  },
  { id: 'created_at', name: t('单据创建时间'), type: 'datetime' },
  { id: 'applied_type', name: t('申请类型'), type: 'enum', option: APPLIED_TYPE_NAME },
  { id: 'applied_core', name: t('申请数（核）'), type: 'number' },
  { id: 'delivered_core', name: t('已交付（核）'), type: 'number' },
  { id: 'returned_way', name: t('退还方式'), type: 'enum', option: RETURNED_WAY_NAME },
  { id: 'applied_record_id', name: t('关联单据'), type: 'string', meta: { search: { op: QueryRuleOPEnum.IN } } },
  { id: 'match_applied_core', name: t('已退还（核）'), type: 'number' },
  { id: 'returned_core', name: t('已退还（核）'), type: 'number' },
  { id: 'not_returned_core', name: t('未退还（核）'), type: 'number' },
  { id: 'exec_rate', name: t('执行率'), type: 'string' },
] as ModelProperty[];
