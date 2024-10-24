import { AppliedType } from '@/store';

export { default } from './column-common';
export const appliedFieldIds = [
  'order_id',
  'bk_biz_id',
  'created_at',
  'applied_type',
  'applied_core',
  'delivered_core',
  'creator',
];
export const appliedViewProperties = [
  { id: 'order_id', name: '单据ID', type: 'string' },
  { id: 'bk_biz_id', name: '业务', type: 'number' },
  { id: 'created_at', name: '单据创建日期', type: 'datetime' },
  { id: 'applied_type', name: '申请类型', type: 'enum', option: AppliedType },
  { id: 'applied_core', name: '申请数（核）', type: 'number' },
  { id: 'delivered_core', name: '已交付（核）', type: 'number' },
  { id: 'creator', name: '申请人', type: 'string' },
];
