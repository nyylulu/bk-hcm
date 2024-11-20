/* eslint-disable @typescript-eslint/no-unused-vars */
import { h } from 'vue';
import { Button } from 'bkui-vue';
import { Model, Column } from '@/decorator';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import routerAction from '@/router/utils/action';
import useSearchQs from '@/hooks/use-search-qs';
import type { IStatsItem } from '@/store/green-channel/stats';
import BusinessValue from '@/components/display-value/business-value.vue';

@Model('green-channel/stats-list.view')
export class StatsListView {
  @Column('business', {
    name: '业务',
    render: ({ data }: { data?: IStatsItem }) =>
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            const searchQs = useSearchQs();
            routerAction.open({
              name: 'ApplicationsManage',
              query: {
                [GLOBAL_BIZS_KEY]: data.bk_biz_id,
                type: 'host_apply',
                initial_filter: searchQs.build({ requireType: [7], bkUsername: [] }),
              },
            });
          },
        },
        h(BusinessValue, { value: data.bk_biz_id }),
      ),
  })
  bk_biz_id: number;

  @Column('number', { name: '单据数量', sort: true, align: 'right' })
  order_count: number;

  @Column('number', { name: '已交付', unit: '核', sort: true, align: 'right' })
  sum_delivered_core: number;

  @Column('number', { name: '申请数', unit: '核', sort: true, align: 'right' })
  sum_applied_core: number;
}
