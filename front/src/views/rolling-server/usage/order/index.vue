<script setup lang="ts">
import { onMounted, ref, useTemplateRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { IRollingServerCpuCoreSummary, RollingServerRecordItem, useRollingServerUsageStore } from '@/store';
import routerAction from '@/router/utils/action';
import { convertDateRangeToObject, getDateRange, transformSimpleCondition } from '@/utils/search';
import { MENU_BUSINESS_ROLLING_SERVER_USAGE_APPLIED } from '@/constants/menu-symbol';
import usageOrderViewProperties from '@/model/rolling-server/usage-order.view';
import { ISearchCondition, IView } from '../typings';

import Search from './children/search.vue';
import DataList from './children/data-list.vue';
import ReturnedRecordsDialog from './children/returned-records-dialog.vue';
import useTimeoutPoll from '@/hooks/use-timeout-poll';

const route = useRoute();
const { getBizsId } = useWhereAmI();
const rollingServerUsageStore = useRollingServerUsageStore();

const searchQs = useSearchQs({ key: 'filter', properties: usageOrderViewProperties });
const { pagination, getPageParams } = usePage();

const docList = ref<RollingServerRecordItem[]>();
const summaryInfo = ref<IRollingServerCpuCoreSummary>();
const condition = ref<Record<string, any>>({});

const recordsPoll = useTimeoutPoll(() => {}, 10000);

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      created_at: getDateRange('last30d'),
      suborder_id: [],
      bk_biz_id: [],
    });
    const { created_at, bk_biz_id, suborder_id } = condition.value;
    const bk_biz_ids = bk_biz_id.length === 1 && bk_biz_id[0] === -1 ? undefined : bk_biz_id;
    const { start, end } = convertDateRangeToObject(created_at);

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    // 请求申请单据列表
    const [listRes, summaryRes] = await Promise.all([
      rollingServerUsageStore.getAppliedRecordList({
        filter: transformSimpleCondition({ ...condition.value, bk_biz_id: bk_biz_ids }, usageOrderViewProperties),
        page: getPageParams(pagination, { sort, order }),
      }),
      rollingServerUsageStore.getCpuCoreSummary({ start, end, bk_biz_ids, suborder_ids: suborder_id }),
    ]);
    const { list: appliedRecordList, count } = listRes;

    // 请求返回记录列表(一个申请单，不会有太多回收单，分页数量最大传500)
    const { list: returnedRecordList } = await rollingServerUsageStore.getReturnedRecordList({
      filter: transformSimpleCondition(
        { applied_record_id: appliedRecordList.map((appliedRecordItem) => appliedRecordItem.id) },
        usageOrderViewProperties,
      ),
      page: getPageParams({ current: 1, limit: 500, count: 0 }),
    });

    const returned_core = returnedRecordList.reduce((acc, cur) => acc + cur.match_applied_core, 0);

    // 设置列表
    docList.value = appliedRecordList.map((appliedRecordItem) => {
      return {
        ...appliedRecordItem,
        returned_records: returnedRecordList,
        // 前端计算字段
        returned_core,
        not_returned_core: appliedRecordItem.delivered_core - returned_core,
        exec_rate: `${(returned_core / appliedRecordItem.delivered_core) * 100}%`,
      };
    });
    // 设置汇总信息
    summaryInfo.value = summaryRes;

    // 设置页面总条数
    pagination.count = count;
  },
  { immediate: true },
);

const handleSearch = (vals: ISearchCondition) => {
  searchQs.set(vals);
};

const handleReset = () => {
  searchQs.clear();
};

// todo：调整route name
const handleViewDetails = (id: string) => {
  routerAction.redirect(
    { name: MENU_BUSINESS_ROLLING_SERVER_USAGE_APPLIED, params: { id }, query: { bizs: getBizsId() } },
    { history: true },
  );
};

const returnedRecordsDialogRef = useTemplateRef<typeof ReturnedRecordsDialog>('returned-records-dialog');
const handleReturnedRecords = (id: string) => {
  returnedRecordsDialogRef.value.show(docList.value.find((item) => item.id === id));
};

onMounted(() => {
  recordsPoll.resume();
});
</script>

<template>
  <search :view="IView.ORDER" :condition="condition" @search="handleSearch" @reset="handleReset" />
  <data-list
    v-bkloading="{ loading: rollingServerUsageStore.appliedRecordsListLoading }"
    :view="IView.ORDER"
    :list="docList"
    :pagination="pagination"
    :summary-info="summaryInfo"
    @view-details="handleViewDetails"
    @show-returned-records="handleReturnedRecords"
  />
  <returned-records-dialog ref="returned-records-dialog" />
</template>
