import { ModelProperty } from '@/model/typings';

export default [
  {
    id: 'bk_biz_id',
    name: '业务ID',
    type: 'number',
  },
  {
    id: 'order_id',
    name: '订单号',
    type: 'string',
  },
  {
    id: 'suborder_id',
    name: '关联单据',
    type: 'string',
    meta: {
      display: {
        appearance: 'link',
      },
    },
  },
  {
    id: 'delivered_core',
    name: '已交付',
    type: 'number',
    unit: '核',
    meta: {
      column: {
        align: 'right',
      },
    },
  },
  {
    id: 'returned_core',
    name: '已退还',
    type: 'number',
    unit: '核',
    meta: {
      column: {
        align: 'right',
      },
    },
  },
  {
    id: 'not_returned_core',
    name: '当天未退还',
    type: 'number',
    unit: '核',
    meta: {
      column: {
        align: 'right',
      },
    },
  },
  {
    id: 'year',
    name: '年',
    type: 'number',
  },
  {
    id: 'month',
    name: '月',
    type: 'number',
  },
  {
    id: 'day',
    name: '日',
    type: 'number',
  },
] as ModelProperty[];
