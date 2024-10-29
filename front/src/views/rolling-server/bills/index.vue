<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import dayjs from 'dayjs';
import { useRoute } from 'vue-router';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useRollingServerBillsStore, type IRollingServerBillItem } from '@/store/rolling-server-bills';
import { transformSimpleCondition } from '@/utils/search';
import billsViewProperties from '@/model/rolling-server/bills.view';
import type { IBillsSearchCondition } from './typings';

import Search from './children/search.vue';
import DataList from './children/data-list.vue';
import DialogFineDetails from './children/dialog-fine-details.vue';

const route = useRoute();
const businessGlobalStore = useBusinessGlobalStore();
const rollingServerBillsStore = useRollingServerBillsStore();
const searchQs = useSearchQs({ key: 'filter', properties: billsViewProperties });
const { pagination, getPageParams } = usePage();

const billList = ref<IRollingServerBillItem[]>([]);
const condition = ref<IBillsSearchCondition>({});

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      date: [dayjs().subtract(29, 'day').toDate(), dayjs().toDate()],
      bk_biz_id: await businessGlobalStore.getFirstBizId(),
    });

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'date') as string;
    const order = (query.order || 'DESC') as string;

    const { list, count } = await rollingServerBillsStore.getBillList({
      filter: transformSimpleCondition(condition.value, billsViewProperties),
      page: getPageParams(pagination, { sort: sort === 'date' ? 'day' : sort, order }),
    });

    // const listx = [
    //   {
    //     id: 'a11111',
    //     bk_biz_id: 2005000002,
    //     offset_config_id: 'ab1323',
    //     product_id: 1233,
    //     delivered_core: 109323,
    //     returned_core: 322,
    //     not_returned_core: 3023,
    //     year: 2024,
    //     month: 10,
    //     day: 29,
    //     creator: 'test',
    //     created_at: '2024-10-02T15:04:05Z',
    //   },
    // ];
    // const newList = listx.map((row) => ({
    //   ...row,
    //   date: `${row.year}-${row.month}-${row.day}`,
    // }));

    billList.value = list;
    // billList.value = newList;

    // 设置页码总条数
    pagination.count = count ?? 100;
    // pagination.count = 100;
  },
  {
    immediate: true,
  },
);

const dialog = reactive({
  isShow: false,
  isHidden: true,
  props: {
    dataRow: null,
  },
});

const handleSearch = (values: IBillsSearchCondition) => {
  searchQs.set(values);
};

const handleReset = () => {
  searchQs.clear();
};

const handleViewFineDetails = (row: IRollingServerBillItem) => {
  dialog.props = { dataRow: row };
  dialog.isShow = true;
  dialog.isHidden = false;
};
</script>

<template>
  <search :condition="condition" @search="handleSearch" @reset="handleReset" />
  <data-list
    v-bkloading="{ loading: rollingServerBillsStore.billListLoading }"
    :list="billList"
    :pagination="pagination"
    @view-fine-details="handleViewFineDetails"
  />

  <template v-if="!dialog.isHidden">
    <dialog-fine-details v-model="dialog.isShow" :data-row="dialog.props.dataRow" @hidden="dialog.isHidden = true" />
  </template>
</template>

<style lang="scss" scoped></style>
