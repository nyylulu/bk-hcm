import { defineComponent, computed } from 'vue';
import Panel from '@/components/panel';
import cssModule from '../index.module.scss';
import { Plus as PlusIcon } from 'bkui-vue/lib/icon';
import { Button } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { getTableNewRowClass } from '@/common/util';

export default defineComponent({
  setup() {
    const totalList = [
      {
        key: 'total_cpu_cores',
        label: 'CPU总核数：',
        value: '123',
      },
      {
        key: 'total_memory',
        label: '内存总量：',
        value: '123',
      },
      {
        key: 'total_cloud_disk',
        label: '云盘总量：',
        value: '123',
      },
    ];
    const renderTableData = [
      {
        project_type: '项目1',
        total_cpu_cores: '123',
      },
    ];

    const { columns, settings } = useColumns('forecastList');
    const tableColumns = computed(() => {
      return [
        ...columns,
        {
          label: '操作',
          width: 120,
          render: () => (
            <div>
              <Button text theme={'primary'} onClick={() => handleToCopy()}>
                克隆
              </Button>
            </div>
          ),
        },
      ];
    });
    const { CommonTable } = useTable({
      searchOptions: {
        disabled: true,
      },
      tableOptions: {
        columns: tableColumns.value,
        reviewData: renderTableData,
        extra: {
          rowClass: getTableNewRowClass(),
          settings: settings.value,
        },
      },
      requestOption: {
        type: '',
      },
    });

    const handleToAdd = () => {};
    const handleToCopy = () => {};
    return () => (
      // 预测清单
      <Panel title='预测清单'>
        <div class={cssModule.flex}>
          <Button theme='primary' outline onClick={() => handleToAdd()} class={cssModule['btn-margin']}>
            <PlusIcon class={cssModule['add-font']} />
            添加
          </Button>
          {totalList.map((item) => {
            return (
              <div class={cssModule['ml-24']}>
                <span>{item.label}</span> <span class={cssModule['total-color']}>{item.value}</span>
              </div>
            );
          })}
        </div>
        <CommonTable></CommonTable>
      </Panel>
    );
  },
});
