import { ModelProperty } from '../typings';
import usageBaseView from './usage-base.view';
import { APPLIED_TYPE_NAME } from '@/views/rolling-server/usage/constants';

export default [
  ...usageBaseView,
  {
    id: 'applied_type',
    name: '申请类型',
    type: 'enum',
    option: APPLIED_TYPE_NAME,
  },
  {
    id: 'applied_core',
    name: '申请数（核）',
    type: 'number',
  },
  {
    id: 'delivered_core',
    name: '已交付（核）',
    type: 'number',
  },
] as ModelProperty[];
