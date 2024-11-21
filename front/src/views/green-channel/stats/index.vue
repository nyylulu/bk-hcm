<script setup lang="ts">
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import dayjs from 'dayjs';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { useGreenChannelStatsStore, type IStatsItem } from '@/store/green-channel/stats';
import { useGreenChannelQuotaStore } from '@/store/green-channel/quota';
import { transformFlatCondition } from '@/utils/search';
import { getModel } from '@/model/manager';
import { StatsSearchView } from '@/model/green-channel/stats-search.view';
import type { ISearchCondition } from './typings';

import Search from './children/search.vue';
import DataList from './children/data-list.vue';

const properties = getModel(StatsSearchView).getProperties();

const route = useRoute();
const greenChannelStatsStore = useGreenChannelStatsStore();
const greenChannelQuotaStore = useGreenChannelQuotaStore();
const searchQs = useSearchQs({ key: 'filter', properties });
const { pagination, getPageParams } = usePage();

const statsList = ref<IStatsItem[]>([]);
const condition = ref<ISearchCondition>({});

const deliveredCore = ref();

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      date: [dayjs().subtract(29, 'day').toDate(), dayjs().toDate()],
    });

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'order_count') as string;
    const order = (query.order || 'DESC') as string;

    const [listRes, cpuCoreRes] = await Promise.allSettled([
      greenChannelStatsStore.getStatsList({
        ...transformFlatCondition(condition.value, properties),
        page: getPageParams(pagination, { sort, order }),
      }),
      greenChannelQuotaStore.getCpuCoreSummary(transformFlatCondition(condition.value, properties)),
    ]);

    if (listRes.status === 'fulfilled') {
      statsList.value = listRes.value.list;
      pagination.count = listRes.value.count;
    }

    if (cpuCoreRes.status === 'fulfilled') {
      deliveredCore.value = cpuCoreRes.value?.sum_delivered_core;
    }
  },
  {
    immediate: true,
  },
);

const handleSearch = (values: ISearchCondition) => {
  searchQs.set(values);
};

const handleReset = () => {
  searchQs.clear();
};
</script>

<template>
  <search :condition="condition" @search="handleSearch" @reset="handleReset" />
  <div class="toolbar">
    <div class="summary">
      <div class="item">
        CPU核-已交付：
        <span class="value">{{ deliveredCore ?? '--' }}</span>
      </div>
    </div>
  </div>
  <data-list
    v-bkloading="{ loading: greenChannelStatsStore.statsListLoading }"
    :list="statsList"
    :pagination="pagination"
  />
</template>

<style lang="scss" scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 32px 0 24px;
  background: #fff;

  .summary {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 40px;

    .item {
      display: inline-flex;
      .value {
        font-weight: 700;
        color: $warning-color;
      }
    }
  }
}
</style>
