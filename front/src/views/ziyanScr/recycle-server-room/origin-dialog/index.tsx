import { defineComponent, computed, PropType, watch } from 'vue';
import ExportToExcelButton from '@/components/export-to-excel-button';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import cssModule from '../dialog.module.scss';
import { useI18n } from 'vue-i18n';

import { IDissolveHostCurrentListParam } from '@/typings/ziyanScr';
import { useTable } from '@/hooks/useResourcePlanTable';
import type { IPageQuery } from '@/typings';
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
      type: Object as PropType<IDissolveHostCurrentListParam>,
    },
  },
  emits: ['update:is-show'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const { columns, settings } = useColumns('decommissionDetails');
    const ziyanScrStore = useZiyanScrStore();

    const title = computed(() =>
      t('{title}（原始）_设备详情', { title: props?.searchParams?.bk_biz_names?.[0] || '' }),
    );

    const getData = (page: IPageQuery) => {
      return ziyanScrStore.dissolveHostOriginList({
        page,
        ...props.searchParams,
      });
    };

    const { tableData, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } =
      useTable(getData, 'details');

    const handleClose = () => {
      emit('update:is-show', false);
    };
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
          <export-to-excel-button data={tableData.value} columns={columns} filename={title.value} theme='primary' />
          <span class={cssModule['total-num']}>
            {t('总条数：')} {pagination.value.count}
          </span>
        </div>
        <bk-loading loading={isLoading.value}>
          <bk-table
            show-overflow-tooltip
            remote-pagination
            data={tableData.value}
            columns={columns}
            pagination={pagination.value}
            settings={settings.value}
            onPageLimitChange={handlePageSizeChange}
            onPageValueChange={handlePageChange}
            onColumnSort={handleSort}
          />
        </bk-loading>
      </bk-dialog>
    );
  },
});
