<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import { Plus } from 'bkui-vue/lib/icon';
import { useRoute } from 'vue-router';
import routeQuery from '@/router/utils/query';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { useUserStore } from '@/store';
import { useRollingServerQuotaStore, type IRollingServerBizQuotaItem } from '@/store/rolling-server-quota';
import { transformFlatCondition } from '@/utils/search';
import quotaBizViewProperties from '@/model/rolling-server/quota-biz.view';
import type { IBizViewSearchCondition } from '../typings';

import Search from './children/search.vue';
import DataList from './children/data-list.vue';
import DialogCreate from './children/dialog-create.vue';
import DialogAdjust from './children/dialog-adjust.vue';
import DialogRecord from './children/dialog-record.vue';

const route = useRoute();
const userStore = useUserStore();
const rollingServerQuotaStore = useRollingServerQuotaStore();
const searchQs = useSearchQs({ key: 'filter', properties: quotaBizViewProperties });
const { pagination, getPageParams } = usePage();

const quotaList = ref<IRollingServerBizQuotaItem[]>([]);
const condition = ref<IBizViewSearchCondition>({});

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      quota_month: new Date(),
      reviser: userStore.username,
    });

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    const { list, count } = await rollingServerQuotaStore.getBizQuotaList({
      ...transformFlatCondition(condition.value, quotaBizViewProperties),
      page: getPageParams(pagination, { sort, order }),
    });

    quotaList.value = list;

    // 设置页码总条数
    pagination.count = count ?? 1;
  },
  {
    immediate: true,
  },
);

const dialog = reactive({
  isShow: false,
  isHidden: true,
  component: null,
  props: {},
});

const handleSearch = (values: IBizViewSearchCondition) => {
  searchQs.set(values);
};

const handleReset = () => {
  searchQs.clear();
};

const handleViewRecord = (row: IRollingServerBizQuotaItem) => {
  dialog.component = DialogRecord;
  dialog.props = { dataRow: row };
  dialog.isShow = true;
  dialog.isHidden = false;
};

const handleAdjust = (row?: IRollingServerBizQuotaItem) => {
  dialog.component = DialogAdjust;
  dialog.props = { dataRow: row };
  dialog.isShow = true;
  dialog.isHidden = false;
};

const handleCreate = () => {
  dialog.component = DialogCreate;
  dialog.isShow = true;
  dialog.isHidden = false;
};

const handleAdjustSuccess = (ids: string[], isCurrent: boolean) => {
  if (isCurrent) {
    routeQuery.refresh();
  }
};
const handleCreateSuccess = () => {
  routeQuery.refresh();
};
</script>

<template>
  <search :condition="condition" @search="handleSearch" @reset="handleReset" />
  <div class="toolbar">
    <bk-button theme="primary" @click="handleCreate">
      <plus style="font-size: 22px" />
      新增额度
    </bk-button>
    <bk-button @click="handleAdjust()">跨月额度调整</bk-button>
  </div>
  <data-list
    v-bkloading="{ loading: rollingServerQuotaStore.bizQuotaListLoading }"
    :list="quotaList"
    :pagination="pagination"
    @view-record="handleViewRecord"
    @adjust="handleAdjust"
  />

  <!-- isHidden为了销毁dialog不保留数据 -->
  <template v-if="!dialog.isHidden">
    <component
      :is="dialog.component"
      v-model="dialog.isShow"
      v-bind="dialog.props"
      @hidden="dialog.isHidden = true"
      @adjust-success="handleAdjustSuccess"
      @create-success="handleCreateSuccess"
    />
  </template>
</template>

<style lang="scss" scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 24px 0 16px 0;
}
</style>
