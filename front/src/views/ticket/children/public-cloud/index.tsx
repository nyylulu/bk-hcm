import { defineComponent, PropType } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { Button } from 'bkui-vue';

import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { searchData } from '../../constants';
import { RulesItem } from '@/typings';
import { MENU_BUSINESS_TICKET_DETAILS } from '@/constants/menu-symbol';

export default defineComponent({
  props: {
    rules: {
      type: Array as PropType<RulesItem[]>,
      required: true,
    },
  },
  setup(props) {
    const router = useRouter();
    const route = useRoute();

    const { columns } = useColumns('myApply');
    const { CommonTable } = useTable({
      searchOptions: {
        searchData,
      },
      tableOptions: {
        columns: [
          {
            label: '单号',
            field: 'sn',
            render: ({ data }: any) => (
              <Button
                text
                theme='primary'
                onClick={() => {
                  router.push({
                    name: MENU_BUSINESS_TICKET_DETAILS,
                    query: { ...route.query, id: data.id, source: data.source },
                  });
                }}>
                {data.sn}
              </Button>
            ),
          },
          ...columns,
        ],
      },
      requestOption: {
        type: 'applications',
        filterOption: { rules: props.rules },
      },
    });

    return () => <CommonTable class={cssModule.table} />;
  },
});
