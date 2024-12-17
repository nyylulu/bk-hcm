import { IDemandListDetail } from '@/typings/plan';
import { formatDisplayNumber } from '@/utils';
import { ExclamationCircleShape } from 'bkui-vue/lib/icon';
import { computed, Ref } from 'vue';

export const useModColumn = (originData: Ref<IDemandListDetail[]>) => {
  const renderDiff = (cur: IDemandListDetail, origin: IDemandListDetail, field: keyof IDemandListDetail) => {
    if (cur[field] === origin?.[field]) return cur[field];
    return (
      <span class={'plan-mod-diff-txt'}>
        <ExclamationCircleShape
          height={14}
          width={14}
          class={'mr4'}
          fill='#FF9C01'
          v-bk-tooltips={{
            content: `修改前：${formatDisplayNumber(origin?.[field])}`,
          }}
        />
        {formatDisplayNumber(cur[field])}
      </span>
    );
  };

  const columns = computed(() => [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    {
      label: '预测ID',
      field: 'demand_id',
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
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'device_type'),
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'expect_time'),
    },
    {
      label: '实例剩余数',
      field: 'remained_os',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'remained_os'),
    },
    {
      label: 'CPU剩余核数',
      field: 'remained_cpu_core',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'remained_cpu_core'),
    },
    {
      label: '内存剩余量(GB)',
      field: 'remained_memory',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'remained_memory'),
    },
    {
      label: '云盘剩余量(GB)',
      field: 'remained_disk_size',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'remained_disk_size'),
    },
    {
      label: '城市',
      field: 'region_name',
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'region_name'),
    },
    {
      label: '可用区',
      field: 'zone_name',
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'zone_name'),
    },
    {
      label: '项目类型',
      field: 'obs_project',
      isDefaultShow: true,
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'obs_project'),
    },
    {
      label: '云磁盘类型',
      field: 'disk_type_name',
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'disk_type_name'),
    },
    {
      label: '单实例磁盘IO(MB/s)',
      field: 'disk_io',
      render: ({ data, index }: { data: IDemandListDetail; index: number }) =>
        renderDiff(data, originData.value[index], 'disk_io'),
    },
  ]);

  return columns.value;
};
