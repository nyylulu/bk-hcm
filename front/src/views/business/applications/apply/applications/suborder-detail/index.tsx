import { defineComponent, PropType, ref, watch } from 'vue';

import { Table } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import CommonDialog from '@/components/common-dialog';

import { useZiyanScrStore } from '@/store';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import usePagination from '@/hooks/usePagination';
import useTimeoutPoll from '@/hooks/use-timeout-poll';

export interface SubOrderInfo {
  step_name: string;
  step_id: number;
  suborder_id: number;
}

export default defineComponent({
  props: {
    subOrderInfo: { type: Object as PropType<SubOrderInfo>, required: true },
  },
  setup(props, { expose }) {
    const DETAIL_STATUS = [
      { label: undefined, name: '全部' },
      { label: 0, name: '成功' },
      { label: 2, name: '失败' },
      { label: 1, name: '执行中' },
    ];
    const scrStore = useZiyanScrStore();

    const isDialogShow = ref(false);
    const list = ref([]);
    const isLoading = ref(false);
    const curStatus = ref();

    const { columns: producingColumns } = useColumns('scrProduction');
    const { columns: initialColumns } = useColumns('scrInitial');
    const { columns: deliveryColumns } = useColumns('scrDelivery');

    const fetchApi = ref<Function>();
    const tableColumns = ref([]);
    const { pagination, handlePageLimitChange, handlePageValueChange } = usePagination(() => getListData());

    const getListData = async () => {
      if (!fetchApi.value) return;
      isLoading.value = true;
      try {
        const { data } = await fetchApi.value(
          props.subOrderInfo.suborder_id,
          {
            limit: pagination.limit,
            start: pagination.start,
          },
          curStatus.value,
        );
        list.value = data.info;
        pagination.count = data?.count;
      } catch (error) {
        list.value = [];
        pagination.count = 0;
      } finally {
        isLoading.value = false;
      }
    };

    const suborderDetailPolling = useTimeoutPoll(
      () => {
        getListData();
      },
      30000,
      { max: 60 },
    );

    watch(
      () => [props.subOrderInfo.suborder_id, props.subOrderInfo.step_id, curStatus.value],
      () => {
        switch (props.subOrderInfo.step_id) {
          case 2: {
            fetchApi.value = scrStore.getProductionDetails;
            tableColumns.value = producingColumns;
            break;
          }
          case 3: {
            fetchApi.value = scrStore.getInitializationDetails;
            tableColumns.value = initialColumns;
            break;
          }
          case 4: {
            fetchApi.value = scrStore.getDeliveryDetails;
            tableColumns.value = deliveryColumns;
            break;
          }
        }
        pagination.start = 0;
        getListData();
      },
      { immediate: true, deep: true },
    );

    watch(
      isDialogShow,
      (isShow) => {
        if (isShow) {
          suborderDetailPolling.resume();
        } else {
          suborderDetailPolling.reset();
        }
      },
      { immediate: true },
    );

    const triggerShow = (v: boolean) => {
      isDialogShow.value = v;
    };

    expose({ triggerShow });

    return () => (
      <CommonDialog
        v-model:isShow={isDialogShow.value}
        title={`资源${props.subOrderInfo.step_name}详情`}
        width={1200}
        dialogType='show'>
        <BkRadioGroup v-model={curStatus.value}>
          {DETAIL_STATUS.map(({ label, name }) => (
            <BkRadioButton label={label}>{name}</BkRadioButton>
          ))}
        </BkRadioGroup>
        <bk-loading loading={isLoading.value}>
          <Table
            style={{ maxHeight: '600px', marginTop: '16px' }}
            data={list.value}
            remotePagination
            pagination={pagination}
            columns={tableColumns.value}
            showOverflowTooltip
            onPageLimitChange={handlePageLimitChange}
            onPageValueChange={handlePageValueChange}
          />
        </bk-loading>
      </CommonDialog>
    );
  },
});
