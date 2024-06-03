import { defineComponent } from 'vue';
import './index.scss';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
export default defineComponent({
  name: 'ResourceConfirm',
  setup() {
    const { columns } = useColumns('DetermineBusiness');
    const { CommonTable } = useTable({
      tableOptions: {
        columns,
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
      slotAllocation: () => {
        return {
          ScrSwitch: false,
          interface: {
            Parameters: {
              filter: {
                condition: 'AND',
                rules: [
                  {
                    field: 'require_type',
                    operator: 'equal',
                    value: 1,
                  },
                  {
                    field: 'label.device_group',
                    operator: 'in',
                    value: ['标准型'],
                  },
                ],
              },
              page: { limit: 0, count: 0 },
            },
            filter: { simpleConditions: true, requestId: 'devices' },
            path: '/api/v1/woa/config/findmany/config/cvm/device/detail',
          },
        };
      },
    });
    return () => (
      <div class='div-ResourceSelect'>
        <CommonTable></CommonTable>
      </div>
    );
  },
});
