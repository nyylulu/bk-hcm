export default [
  {
    id: 'id',
    name: 'ID',
    type: 'string',
  },
  {
    id: 'bk_biz_id',
    name: '业务',
    type: 'bizs',
  },
  {
    id: 'order_id',
    name: '单据ID',
    type: 'array',
  },
  {
    id: 'suborder_id',
    name: '子单据ID',
    type: 'string',
  },
  {
    id: 'year',
    name: '申请时间年份',
    type: 'string',
  },
  {
    id: 'month',
    name: '申请时间月份',
    type: 'string',
  },
  {
    id: 'day',
    name: '申请时间天',
    type: 'string',
  },
  {
    id: 'creator',
    name: '创建者',
    type: 'user',
  },
  {
    id: 'created_at',
    name: '单据创建时间',
    type: 'datetime',
  },
];
