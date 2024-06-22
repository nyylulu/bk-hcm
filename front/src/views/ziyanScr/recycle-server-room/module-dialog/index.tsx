import { defineComponent, computed, PropType, watch, ref } from 'vue';
import ExportToExcelButton from '@/components/export-to-excel-button';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import cssModule from '../dialog.module.scss';
import { useI18n } from 'vue-i18n';
import { useZiyanScrStore } from '@/store/ziyanScr';
import { IDissolveHostOriginListParam } from '@/typings/ziyanScr';

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

    const isLoading = ref(false);
    const tableData = ref();
    const pagination = ref({
      current: 1,
      limit: 10,
      count: 0,
    });

    const title = computed(() => t('{title}_设备详情', { title: props?.searchParams?.bk_biz_names?.[0] || '' }));

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
            columns={columns}
            pagination={pagination.value}
            settings={settings.value}
          />
        </bk-loading>
      </bk-dialog>
    );
  },
});
