import { ModelProperty } from '@/model/typings';
import { quotaAdjustTypeNames } from '@/views/rolling-server/constants';

export default [
  {
    id: 'quota_month',
    name: '额度月份',
    type: 'datetime',
  },
  {
    id: 'bk_biz_ids',
    name: '业务',
    type: 'business',
  },
  {
    id: 'bk_biz_name',
    name: '业务',
    type: 'string',
  },
  {
    id: 'adjust_type',
    name: '调整类型',
    type: 'enum',
    option: quotaAdjustTypeNames,
  },
  {
    id: 'reviser',
    name: '更新人',
    type: 'user',
  },
  {
    id: 'base_quota',
    name: '基础额度',
    type: 'number',
  },
  {
    id: 'quota_offset',
    name: '调整额度',
    type: 'number',
  },
  {
    id: 'quota_offset_final',
    name: '调整后实际额度',
    type: 'number',
  },
  {
    id: 'updated_at',
    name: '更新时间',
    type: 'datetime',
  },
  {
    id: 'created_at',
    name: '时间',
    type: 'datetime',
  },
] as ModelProperty[];
