import { Ref, ref } from 'vue';
import { RouteLocationNormalizedLoaded } from 'vue-router';

import BillsExportButton from '@/views/bill/bill/components/bills-export-button';
import BccSyncButton from '@/views/bill/bill/summary/primary/bcc-sync-button';

import { useI18n } from 'vue-i18n';
import { reqBillsProductSummaryList, exportBillsProductSummary } from '@/api/bill';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { BillSearchRules } from '@/utils';
import { BILL_MAIN_ACCOUNTS_KEY } from '@/constants';
import { PluginHandlerType } from '../bill-manage';

// 账单汇总-一级账号
const usePrimaryHandler = () => {
  const renderOperation = (bill_year: number, bill_month: number) => {
    return <BccSyncButton billYear={bill_year} billMonth={bill_month} />;
  };

  return {
    renderOperation,
  };
};

// 账单汇总-二级账号
const useSubHandler = () => {
  // mounted 时, 根据初始条件加载表格数据
  const mountedCallback = (route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
    // 只有二级账号有保存的需求
    const billSearchRules = new BillSearchRules();
    billSearchRules.addRule(route, BILL_MAIN_ACCOUNTS_KEY, 'main_account_id', QueryRuleOPEnum.IN);
    reloadTable(billSearchRules.rules);
  };

  return {
    mountedCallback,
  };
};

// 账单汇总-业务
const useProductHandler = () => {
  // table 相关状态
  const selectedIds = ref<number[]>([]);
  const columnName = 'billsProductSummary';
  const getColumns = (columns: any[]) => columns;
  const apiMethod: (...args: any) => Promise<any> = reqBillsProductSummaryList;
  const extensionKey = 'op_product_ids';

  // reloadTable 时, 重置选中项
  const reloadSelectedIds = (rules: RulesItem[]) => {
    // 运营产品这里 rules 只会有一个搜索条件, 直接按索引取就行
    selectedIds.value = rules.length > 0 ? (rules[0].value as number[]) : [];
  };

  // mounted 时, 根据初始条件加载表格数据
  const mountedCallback = (_route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
    reloadTable([]);
  };

  // 操作栏
  const renderOperation = (bill_year: number, bill_month: number, searchRef: Ref<any>) => {
    const { t } = useI18n();

    return (
      <BillsExportButton
        cb={() =>
          exportBillsProductSummary({
            bill_year,
            bill_month,
            export_limit: 200000,
            op_product_ids: searchRef.value.rules.find((rule: any) => rule.field === 'product_id')?.value || [],
          })
        }
        title={t('账单汇总-运营产品')}
        content={t('导出当月运营产品的账单数据')}
      />
    );
  };

  return {
    selectedIds,
    columnName,
    getColumns,
    extensionKey,
    apiMethod,
    reloadSelectedIds,
    mountedCallback,
    renderOperation,
  };
};

// 账单调整
const useAdjustHandler = () => {
  // mounted 时, 根据初始条件加载表格数据
  const mountedCallback = (route: RouteLocationNormalizedLoaded, reloadTable: (rules: RulesItem[]) => void) => {
    // 只有二级账号有保存的需求
    const billSearchRules = new BillSearchRules();
    billSearchRules.addRule(route, BILL_MAIN_ACCOUNTS_KEY, 'main_account_id', QueryRuleOPEnum.IN);
    reloadTable(billSearchRules.rules);
  };

  return {
    mountedCallback,
  };
};

const pluginHandler: PluginHandlerType = {
  usePrimaryHandler,
  useSubHandler,
  useProductHandler,
  useAdjustHandler,
};

export default pluginHandler;
