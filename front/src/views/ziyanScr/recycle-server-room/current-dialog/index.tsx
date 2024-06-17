import { defineComponent, computed, PropType, watch } from 'vue';
import ExportToExcelButton from '@/components/export-to-excel-button';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import cssModule from '../dialog.module.scss';
import { useI18n } from 'vue-i18n';

import type { IDissolveHostOriginListParam } from '@/typings/ziyanScr';
import { useZiyanScrStore } from '@/store/ziyanScr';
import type { IPageQuery } from '@/typings';
import { useTable } from '@/hooks/useResourcePlanTable';

export default defineComponent({
  components: {
    ExportToExcelButton,
  },
  props: {
    isShow: {
      type: Boolean,
    },
    searchParams: {
      type: Object as PropType<IDissolveHostOriginListParam>,
    },
  },
  emits: ['update:is-show'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const { columns, settings } = useColumns('decommissionDetails');
    const ziyanScrStore = useZiyanScrStore();

    const expandMap = {
      device_layer: t('设备技术分类'),
      cpu_score: t('CPU得分'),
      mem_score: t('内存得分'),
      inner_net_traffic_score: t('内网流量得分'),
      disk_io_score: t('磁盘IO得分'),
      disk_util_score: t('磁盘IO使用率得分'),
      is_pass: t('是否达标'),
      mem4linux: t('内存使用量(G)'),
      inner_net_traffic: t('内网流量(Mb/s)'),
      outer_net_traffic: t('外网流量(Mb/s)'),
      disk_io: t('磁盘IO(Blocks/s)'),
      disk_util: t('磁盘IO使用率'),
      disk_total: t('磁盘总量(G)'),
    };

    const title = computed(() =>
      t('{title}_总数（当前）_设备详情', { title: props?.searchParams?.bk_biz_names?.[0] || '' }),
    );
    const tableCoumn = computed(() => {
      return [
        {
          type: 'expand',
          width: 32,
          minWidth: 32,
          colspan: 1,
          resizable: false,
        },
        ...columns,
      ];
    });

    const handleClose = () => {
      emit('update:is-show', false);
    };

    const getData = (page: IPageQuery) => {
      return ziyanScrStore.dissolveHostCurrentList({
        page,
        ...props.searchParams,
      });
    };

    const { tableData, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } =
      useTable(getData, 'details');

    watch(
      () => props.searchParams,
      () => {
        if (props.isShow) {
          triggerApi();
        }
      },
    );

    return () => (
      <bk-dialog
        class='step-dialog'
        width={'80%'}
        dialog-type='show'
        theme='primary'
        headerAlign='left'
        title={title.value}
        isShow={props.isShow}
        onClosed={() => handleClose()}>
        <div class={cssModule.title}>
          <export-to-excel-button data={[]} columns={[]} filename='' />
          <span class={cssModule['total-num']}>{t('总条数：')}</span>
        </div>
        <bk-loading loading={isLoading.value}>
          <bk-table
            show-overflow-tooltip
            remote-pagination
            data={tableData.value}
            columns={tableCoumn.value}
            pagination={pagination.value}
            settings={settings.value}
            onPageLimitChange={handlePageSizeChange}
            onPageValueChange={handlePageChange}
            onColumnSort={handleSort}>
            {{
              expandRow: (row: any) => {
                return (
                  <div class={cssModule['expand-row']}>
                    {Object.keys(expandMap).map((item) => {
                      return (
                        <div class={cssModule['expand-item']}>
                          <span>{expandMap[item] || '--'}</span> {row[item] || '--'}
                        </div>
                      );
                    })}
                  </div>
                );
              },
            }}
          </bk-table>
        </bk-loading>
      </bk-dialog>
    );
  },
});
