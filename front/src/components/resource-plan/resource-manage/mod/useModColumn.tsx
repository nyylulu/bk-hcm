import { IDemandListDetail } from '@/typings/plan';
import { ExclamationCircleShape } from 'bkui-vue/lib/icon';
import { computed, Ref } from 'vue';

export const useModColumn = (originData: Ref<IDemandListDetail[]>) => {
  const renderDiff = (cur: IDemandListDetail, origin: IDemandListDetail, field: keyof IDemandListDetail) => {
    if (cur[field] === origin?.[field]) return cur[field];
    else
      return (
        <span class={'plan-mod-diff-txt'}>
          <ExclamationCircleShape
            height={14}
            width={14}
            class={'mr4'}
            fill='#FF9C01'
            v-bk-tooltips={{
              content: `修改前：${origin?.[field]}`,
            }}
          />
          {cur[field]}
        </span>
      );
  };

  const columns = computed(() => [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    {
      label: '预测ID',
      field: 'crp_demand_id',
      isDefaultShow: true,
    },
    {
      label: '类型',
      field: 'demand_class',
      isDefaultShow: true,
    },
    {
      label: '机型规格',
      field: 'device_type',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'device_type'),
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'expect_time'),
    },
    {
      label: '实例总数',
      field: 'total_os',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'total_os'),
    },
    {
      label: 'CPU总核数',
      field: 'total_cpu_core',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'total_cpu_core'),
    },
    {
      label: '内存总量(GB)',
      field: 'total_memory',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'total_memory'),
    },
    {
      label: '云盘总量(GB)',
      field: 'total_disk_size',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'total_disk_size'),
    },
    {
      label: '城市',
      field: 'region_name',
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'region_name'),
    },
    {
      label: '可用区',
      field: 'zone_name',
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'zone_name'),
    },
    {
      label: '项目类型',
      field: 'obs_project',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'obs_project'),
    },
    {
      label: '云磁盘类型',
      field: 'disk_type_name',
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'disk_type_name'),
    },
    {
      label: '单实例磁盘IO(MB/s)',
      field: 'disk_io',
      render: ({ data, index }: { data: IDemandListDetail; index: number }) => renderDiff(data, originData.value[index], 'disk_io'),
    },
  ]);

  return columns.value;
};
