/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';

@Model('green-channel/stats-list.view')
export class StatsListView {
  @Column('business', { name: '业务' })
  bk_biz_id: number;

  @Column('number', { name: '单据数量', sort: true, align: 'right' })
  order_count: number;

  @Column('number', { name: '已交付', unit: '核', sort: true, align: 'right' })
  sum_delivered_core: number;

  @Column('number', { name: '申请数', unit: '核', sort: true, align: 'right' })
  sum_applied_core: number;
}
