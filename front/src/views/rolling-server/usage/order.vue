<script setup lang="ts">
import { ref, useTemplateRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { RollingServerRecordItem, useRollingServerStore } from '@/store';
import routerAction from '@/router/utils/action';
import { getDateRange, transformSimpleCondition } from '@/utils/search';
import { MENU_BUSINESS_ROLLING_SERVER_USAGE_APPLIED } from '@/constants/menu-symbol';
import usageAppliedView from '@/model/rolling-server/usage-applied.view';
import usageReturnedView from '@/model/rolling-server/usage-returned.view';
import { ISearchCondition, IView } from './typings';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import ReturnedRecordsDialog from './children/returned-records-dialog/returned-records-dialog.vue';

import randomRecords from './random-records';

const route = useRoute();
const { getBizsId } = useWhereAmI();
const rollingServerStore = useRollingServerStore();

const usageView = [...usageAppliedView, ...usageReturnedView];
const searchQs = useSearchQs({ key: 'filter', properties: [] });
const { pagination, getPageParams } = usePage();

const docList = ref<RollingServerRecordItem[]>(randomRecords);
const condition = ref<Record<string, any>>({});

const refreshRecords = async () => {};

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      created_at: getDateRange('last30d'),
    });

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;
    transformSimpleCondition(condition.value, usageView);

    // 请求申请单据列表
    const { list: appliedRecordList, count } = await rollingServerStore.getAppliedRecordList({
      filter: transformSimpleCondition(condition.value, usageView),
      page: getPageParams(pagination, { sort, order }),
    });

    // 请求返回记录列表(一个申请单，不会有太多回收单，分页数量最大传500)
    const { list: returnedRecordList } = await rollingServerStore.getReturnedRecordList({
      filter: transformSimpleCondition(
        {
          // todo: 确定对应条件
          applied_record_id: appliedRecordList.map((appliedRecordItem) => appliedRecordItem.suborder_id),
        },
        usageView,
      ),
      page: getPageParams({ current: 1, limit: 500, count: 0 }, { sort, order }),
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

    // 设置页面总条数
    pagination.count = count;

    // TODO: 每10s刷新列表
    refreshRecords();
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
</script>

<template>
  <search :view="IView.ORDER" :condition="condition" @search="handleSearch" @reset="handleReset" />
  <data-list
    :view="IView.ORDER"
    :list="docList"
    :pagination="pagination"
    @view-details="handleViewDetails"
    @show-returned-records="handleReturnedRecords"
  />
  <returned-records-dialog ref="returned-records-dialog" />
</template>
