<script setup lang="ts">
import { ref } from 'vue';

import routerAction from '@/router/utils/action';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { useWhereAmI } from '@/hooks/useWhereAmI';

import { appliedProperties } from '@/model/rolling-server/usage/properties';
import { MENU_BUSINESS_ROLLING_SERVER_USAGE_APPLIED } from '@/constants/menu-symbol';
import { ISearchCondition, IView } from './typings';
import { IRollingServerAppliedRecordsItem } from '@/store';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';

const { getBizsId } = useWhereAmI();

const searchQs = useSearchQs({ key: 'filter', properties: appliedProperties });
const { pagination } = usePage();

const appliedList = ref<IRollingServerAppliedRecordsItem[]>([]);
const condition = ref<Record<string, any>>({});

const handleSearch = (vals: ISearchCondition) => {
  searchQs.set(vals);
};

const handleReset = () => {
  searchQs.clear();
};

const handleViewDetails = (id: string) => {
  routerAction.redirect(
    {
      name: MENU_BUSINESS_ROLLING_SERVER_USAGE_APPLIED,
      params: { id },
      query: { bizs: getBizsId() },
    },
    {
      history: true,
    },
  );
};
</script>

<template>
  <search :view="IView.APPLIED" :condition="condition" @search="handleSearch" @reset="handleReset" />
  <data-list :view="IView.APPLIED" :list="appliedList" :pagination="pagination" @view-details="handleViewDetails" />
</template>
