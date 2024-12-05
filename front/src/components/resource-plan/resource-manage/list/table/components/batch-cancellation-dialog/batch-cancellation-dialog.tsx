import { Dialog, Message } from 'bkui-vue';
import { computed, defineComponent, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import cssModule from './index.module.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { IListResourcesDemandsItem } from '@/typings/resourcePlan';
import { useResourcePlanStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';

export default defineComponent({
  props: {
    isShow: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    data: {
      type: Array as PropType<IListResourcesDemandsItem[]>,
      default: () => [] as IListResourcesDemandsItem[],
    },
    handleConfirm: {
      type: Function as PropType<() => void>,
    },
  },
  emits: ['update:isShow', 'refresh'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const { columns, settings } = useColumns('resourceForecastBatchCancel');
    const { cancelResourcesDemands } = useResourcePlanStore();
    const { getBizsId } = useWhereAmI();

    const pagination = computed(() => ({ count: props.data.length, limit: 10 }));
    const totalData = computed(() => {
      const totals = props.data.reduce(
        (pre, item) => {
          pre.totalCpuCores += item.total_cpu_core || 0;
          pre.totalMemoryGB += item.total_memory || 0;
          pre.totalCloudDiskGB += item.total_disk_size || 0;
          return pre;
        },
        { totalCpuCores: 0, totalMemoryGB: 0, totalCloudDiskGB: 0 },
      );
      return {
        table: props.data || [],
        ...totals,
      };
    });

    const handleConfirm = async () => {
      const params = {
        cancel_demands: props.data.map(({ demand_id, remained_cpu_core }) => ({ demand_id, remained_cpu_core })),
      };
      await cancelResourcesDemands(getBizsId(), params);
      Message({ theme: 'success', message: '批量删除成功' });
      emit('update:isShow', false);
      emit('refresh');
    };

    return () => (
      <Dialog
        isShow={props.isShow}
        title={t('批量取消')}
        onClosed={() => emit('update:isShow', false)}
        onConfirm={handleConfirm}
        width={'80%'}>
        <div class={cssModule.warning}>{t('将取消以下的资源预测需求，请确认')}</div>
        <bk-table
          row-hover='auto'
          show-overflow-tooltip
          data={totalData.value.table}
          pagination={pagination.value}
          columns={columns}
          settings={settings.value}
        />
        <div class={cssModule.statistics}>
          <span>
            {t('CPU总核数')}：<span class={cssModule.num}>{totalData.value?.totalCpuCores}</span>
          </span>
          <span>
            {t('内存总量(GB)')}：<span class={cssModule.num}>{totalData.value?.totalMemoryGB}</span>
          </span>
          <span>
            {t('云盘总量(GB)')}：<span class={cssModule.num}>{totalData.value?.totalCloudDiskGB}</span>
          </span>
        </div>
      </Dialog>
    );
  },
});
