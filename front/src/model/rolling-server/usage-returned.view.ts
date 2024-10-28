import { QueryRuleOPEnum } from '@/typings';
import { ModelProperty } from '../typings';
import usageBaseView from './usage-base.view';
import { RETURNED_WAY_NAME } from '@/views/rolling-server/usage/constants';

export default [
  ...usageBaseView,
  {
    id: 'returned_way',
    name: '退还方式',
    type: 'enum',
    option: RETURNED_WAY_NAME,
  },
  {
    id: 'applied_record_id',
    name: '关联单据',
    type: 'string',
    meta: { search: { op: QueryRuleOPEnum.IN } },
  },
  {
    id: 'match_applied_core',
    name: '已退还（核）',
    type: 'number',
  },
  {
    id: 'returned_core',
    name: '已退还（核）',
    type: 'number',
  },
  {
    id: 'not_returned_core',
    name: '未退还（核）',
    type: 'number',
  },
  {
    id: 'exec_rate',
    name: '执行率',
    type: 'string',
  },
] as ModelProperty[];
