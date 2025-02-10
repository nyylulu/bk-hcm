import { defineComponent, computed, PropType, watch, ref } from 'vue';
import ExportToExcelButton from '@/components/export-to-excel-button';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import cssModule from '../dialog.module.scss';
import { useI18n } from 'vue-i18n';

import type { IDissolve, IDissolveHostOriginListParam } from '@/typings/ziyanScr';
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
    rowData: {
      type: Object as PropType<IDissolve>,
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
    const totalCount = ref(0);
    const activeModule = ref('');

    const title = computed(() =>
      t('{title}（当前）_设备详情', { title: props?.searchParams?.bk_biz_names?.[0] || '' }),
    );
    const tableColumn = computed(() => {
      return [
        {
          type: 'expand',
          width: 52,
          minWidth: 52,
          colspan: 1,
          resizable: false,
        },
        ...columns,
      ];
    });

    const handleClose = () => {
      emit('update:is-show', false);
    };

    const handleModuleSearch = (module?: string) => {
      activeModule.value = module;
      getData(module ? [module] : undefined);
    };

    const getData = async (moduleNames?: string[]) => {
      try {
        isLoading.value = true;
        const res = await ziyanScrStore.dissolveHostCurrentList({
          page: {
            count: false,
            start: 0,
            limit: 10000,
          },
          ...props.searchParams,
          module_names: moduleNames ?? props.searchParams.module_names,
        });
        tableData.value = res.data.details;
        pagination.value.count = tableData.value?.length || 0;

        if (!moduleNames) {
          totalCount.value = pagination.value.count;
        }
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
          activeModule.value = '';
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
        <div class={cssModule['dialog-content']}>
          <div class={cssModule['module-list']}>
            <div
              class={[cssModule['module-item'], { [cssModule.active]: !activeModule.value }]}
              onClick={() => handleModuleSearch()}>
              全部 <em class={cssModule['item-count']}>{totalCount.value}</em>
            </div>
            {Object.entries(props?.rowData?.module_host_count || {}).map(([name, count]) => (
              <div
                key={name}
                class={[cssModule['module-item'], { [cssModule.active]: activeModule.value === name }]}
                onClick={() => handleModuleSearch(name)}>
                {name} <em class={cssModule['item-count']}>{count}</em>
              </div>
            ))}
          </div>
          <div class={cssModule['data-table']} v-bkloading={{ loading: isLoading.value }}>
            <div class={cssModule.title}>
              <export-to-excel-button data={tableData.value} columns={columns} filename={title.value} theme='primary' />
              <span class={cssModule['total-num']}>
                {t('总条数：')} {pagination.value.count}
              </span>
            </div>
            <bk-table
              show-overflow-tooltip
              data={tableData.value}
              columns={tableColumn.value}
              pagination={pagination.value}
              settings={settings.value}
              max-height={'calc(100% - 52px)'}>
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
          </div>
        </div>
      </bk-dialog>
    );
  },
});
