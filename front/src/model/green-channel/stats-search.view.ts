/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { convertDateRangeToObject } from '@/utils/search';

@Model('green-channel/stats-search.view')
export class StatsSearchView {
  @Column('business', { name: '业务' })
  bk_biz_ids: number[];

  @Column('datetime', {
    name: '日期范围',
    converter(value) {
      return convertDateRangeToObject(value);
    },
  })
  date: [Date, Date];
}
