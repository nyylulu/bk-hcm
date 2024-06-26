import { defineComponent, computed, PropType, watch, ref } from 'vue';
import ExportToExcelButton from '@/components/export-to-excel-button';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import cssModule from '../dialog.module.scss';
import { useI18n } from 'vue-i18n';

import type { IDissolveHostOriginListParam } from '@/typings/ziyanScr';
import { useZiyanScrStore } from '@/store/ziyanScr';

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

    const isLoading = ref(false);
    const tableData = ref();
    const pagination = ref({
      current: 1,
      limit: 10,
      count: 0,
    });

    const title = computed(() =>
      t('{title}（当前）_设备详情', { title: props?.searchParams?.bk_biz_names?.[0] || '' }),
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

    const getData = async () => {
      try {
        isLoading.value = true;
        const res = await ziyanScrStore.dissolveHostCurrentList({
          page: {
            count: false,
            start: 0,
            limit: 10000,
          },
          ...props.searchParams,
        });
        tableData.value = res.data.details;
        pagination.value.count = tableData.value?.length || 0;
      } catch (error) {
        console.error(error, 'error'); // eslint-disable-line no-console
      } finally {
        isLoading.value = false;
      }
    };

    watch(
      () => props.searchParams,
      () => {
        if (props.isShow) {
          getData();
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
          <export-to-excel-button data={tableData.value} columns={columns} filename={title.value} theme='primary' />
          <span class={cssModule['total-num']}>
            {t('总条数：')} {pagination.value.count}
          </span>
        </div>
        <bk-loading loading={isLoading.value}>
          <bk-table
            show-overflow-tooltip
            data={tableData.value}
            columns={tableCoumn.value}
            pagination={pagination.value}
            settings={settings.value}>
            {{
              expandRow: (row: any) => {
                return (
                  <div class={cssModule['expand-row']}>
                    {Object.keys(expandMap).map((item) => {
                      return (
                        <div class={cssModule['expand-item']}>
                          <span>{expandMap[item] || '--'}</span>
                          {item === 'is_pass' ? t(row[item] ? '达标' : '不达标') : row[item]}
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
