import { ModelProperty } from '@/model/typings';
import { quotaAdjustTypeNames } from '@/views/rolling-server/constants';

export default [
  {
    id: 'adjust_type',
    name: '调整类型',
    type: 'enum',
    option: quotaAdjustTypeNames,
  },
  {
    id: 'operator',
    name: '操作人',
    type: 'user',
  },
  {
    id: 'quota_offset',
    name: '调整额度',
    type: 'number',
  },
  {
    id: 'created_at',
    name: '时间',
    type: 'datetime',
  },
] as ModelProperty[];
