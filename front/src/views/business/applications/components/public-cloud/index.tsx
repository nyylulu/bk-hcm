import { defineComponent, PropType } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import cssModule from './index.module.scss';

import { Button } from 'bkui-vue';

import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { searchData } from '@/views/service/apply-list/constants';
import { RulesItem } from '@/typings';

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
                    path: '/business/applications/detail',
                    query: { ...route.query, id: data.id },
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
