import { Ref, defineComponent, inject, onMounted, ref, watch } from 'vue';

import Button from '../../components/button';
import Amount from '../../components/amount';
import Search from '../../components/search';

import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { reqBillsMainAccountSummarySum, reqBillsProductSummaryList } from '@/api/bill';
import { RulesItem } from '@/typings';

export default defineComponent({
  name: 'OperationProductTabPanel',
  setup() {
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const searchRef = ref();
    const amountRef = ref();

    const op_product_ids = ref<number[]>([]);
    const { columns } = useColumns('billsProductSummary');
    const { CommonTable, getListData, clearFilter, filter } = useTable({
      searchOptions: { disabled: true },
      tableOptions: { columns },
      requestOption: {
        sortOption: {
          sort: 'current_month_rmb_cost',
          order: 'DESC',
        },
        apiMethod: reqBillsProductSummaryList,
        extension: () => ({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
          op_product_ids: op_product_ids.value,
          filter: undefined,
        }),
        immediate: false,
      },
    });

    const reloadTable = (rules: RulesItem[]) => {
      // 运营产品这里 rules 只会有一个搜索条件, 直接按索引取就行
      op_product_ids.value = rules.length > 0 ? (rules[0].value as number[]) : [];
      clearFilter();
      getListData(rules);
    };

    watch([bill_year, bill_month], () => {
      searchRef.value.handleSearch();
    });

    watch(filter, () => {
      amountRef.value.refreshAmountInfo();
    });

    onMounted(() => {
      reloadTable([]);
    });

    return () => (
      <>
        <Search ref={searchRef} searchKeys={['product_id']} onSearch={reloadTable} />
        <div class='p24' style={{ height: 'calc(100% - 162px)' }}>
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
