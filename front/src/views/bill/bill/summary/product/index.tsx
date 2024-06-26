import { Ref, defineComponent, inject, ref, watch } from 'vue';
import Button from '../../components/button';
import Amount from '../../components/amount';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { reqBillsMainAccountSummaryList, reqBillsMainAccountSummarySum } from '@/api/bill';
import { useOperationProducts } from '@/hooks/useOperationProducts';

export default defineComponent({
  name: 'OperationProductTabPanel',
  setup() {
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');
    const amountRef = ref();
    const { getTranslatorMap } = useOperationProducts();

    const { columns } = useColumns('billsMainAccountSummary');
    const { CommonTable, getListData } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns: columns.slice(2, -1),
      },
      requestOption: {
        apiMethod: reqBillsMainAccountSummaryList,
        extension: () => ({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
        }),
        async resolveDataListCb(dataList: any) {
          if (!dataList.length) return;
          const ids = dataList.map((item: { product_id: number }) => item.product_id);
          const map = await getTranslatorMap(ids);
          return dataList.map((data: { product_id: number }) => {
            const { product_id } = data;
            return {
              ...data,
              product_name: map.get(product_id),
            };
          });
        },
      },
    });

    watch([bill_year, bill_month], () => {
      getListData();
      amountRef.value.refreshAmountInfo();
    });

    return () => (
      <div class='full-height p24'>
        <CommonTable>
          {{
            operation: () => <Button noSyncBtn />,
            operationBarEnd: () => (
              <Amount
                ref={amountRef}
                api={reqBillsMainAccountSummarySum}
                payload={() => ({ bill_year: bill_year.value, bill_month: bill_month.value })}
              />
            ),
          }}
        </CommonTable>
      </div>
    );
  },
});
