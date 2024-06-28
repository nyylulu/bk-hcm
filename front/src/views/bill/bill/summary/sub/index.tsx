import { Ref, defineComponent, inject, ref, watch } from 'vue';
import Search from '../../components/search';
import Button from '../../components/button';
import Amount from '../../components/amount';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { reqBillsMainAccountSummaryList, reqBillsMainAccountSummarySum } from '@/api/bill';
import { RulesItem } from '@/typings';
import { useOperationProducts } from '@/hooks/useOperationProducts';

export default defineComponent({
  name: 'SubAccountTabPanel',
  setup() {
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const searchRef = ref();
    const amountRef = ref();
    const { getTranslatorMap } = useOperationProducts();

    const { columns } = useColumns('billsMainAccountSummary');
    const { CommonTable, getListData, clearFilter, filter } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns,
      },
      requestOption: {
        sortOption: {
          sort: 'current_month_rmb_cost',
          order: 'DESC',
        },
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
        immediate: false,
      },
    });

    const reloadTable = (rules: RulesItem[]) => {
      clearFilter();
      getListData(rules);
    };

    watch([bill_year, bill_month], () => {
      searchRef.value.handleSearch();
    });

    watch(filter, () => {
      amountRef.value.refreshAmountInfo();
    });

    return () => (
      <>
        <Search
          ref={searchRef}
          searchKeys={['vendor', 'root_account_id', 'product_id', 'main_account_id']}
          onSearch={reloadTable}
          autoSelectMainAccount
        />
        <div class='p24' style={{ height: 'calc(100% - 238px)' }}>
          <CommonTable>
            {{
              operation: () => <Button noSyncBtn />,
              operationBarEnd: () => (
                <Amount
                  ref={amountRef}
                  api={reqBillsMainAccountSummarySum}
                  payload={() => ({ bill_year: bill_year.value, bill_month: bill_month.value, filter })}
                />
              ),
            }}
          </CommonTable>
        </div>
      </>
    );
  },
});
