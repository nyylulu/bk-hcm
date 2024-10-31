import { useZiyanScrStore } from '@/store';
import { Table } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { defineComponent, ref, watch } from 'vue';
import { DETAIL_STATUS } from './constants';
import './index.scss';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import usePagination from '@/hooks/usePagination';
export default defineComponent({
  props: {
    suborderId: {
      required: true,
      type: Number,
    },
    stepId: {
      required: true,
      type: Number,
    },
  },
  setup(props) {
    const list = ref([]);
    const isLoading = ref(false);
    const scrStore = useZiyanScrStore();
    const curStatus = ref();
    const { columns: producingColumns } = useColumns('scrProduction');
    const { columns: initialColumns } = useColumns('scrInitial');
    const { columns: deliveryColumns } = useColumns('scrDelivery');

    const fetchData = ref<Function>();
    const tableColumns = ref([]);
    const { pagination, handlePageLimitChange, handlePageValueChange } = usePagination(() => getListData());
    const getListData = async () => {
      isLoading.value = true;
      try {
        const { data } = await fetchData.value(
          props.suborderId,
          { limit: pagination.limit, start: pagination.start },
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
    watch(
      () => [props.suborderId, props.stepId, curStatus.value],
      () => {
        switch (props.stepId) {
          case 2: {
            fetchData.value = scrStore.getProductionDetails;
            tableColumns.value = producingColumns;
            break;
          }
          case 3: {
            fetchData.value = scrStore.getInitializationDetails;
            tableColumns.value = initialColumns;
            break;
          }
          case 4: {
            fetchData.value = scrStore.getDeliveryDetails;
            tableColumns.value = deliveryColumns;
            break;
          }
        }
        pagination.start = 0;
        getListData();
      },
      { immediate: true, deep: true },
    );

    return () => (
      <div style={'width: 100%;'}>
        <BkRadioGroup v-model={curStatus.value}>
          {DETAIL_STATUS.map(({ label, name }) => (
            <BkRadioButton label={label}>{name}</BkRadioButton>
          ))}
        </BkRadioGroup>
        <bk-loading loading={isLoading.value}>
          <Table
            data={list.value}
            remotePagination
            pagination={pagination}
            columns={tableColumns.value}
            onPageLimitChange={handlePageLimitChange}
            onPageValueChange={handlePageValueChange}
            class={'maxheigth tablelist'}
          />
        </bk-loading>
      </div>
    );
  },
});
