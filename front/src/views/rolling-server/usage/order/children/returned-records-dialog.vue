<script setup lang="ts">
import { h, ref } from 'vue';
import { Button } from 'bkui-vue';
import { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import usageOrderViewProperties from '@/model/rolling-server/usage-order.view';
import { RollingServerRecordItem } from '@/store';
import useSearchQs from '@/hooks/use-search-qs';
import routerAction from '@/router/utils/action';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { MENU_BUSINESS_TICKET_MANAGEMENT } from '@/constants/menu-symbol';

const isShow = ref(false);
const data = ref<RollingServerRecordItem>();
const show = (row: RollingServerRecordItem) => {
  isShow.value = true;
  data.value = row;
};
const fieldIds = ['created_at', 'match_applied_core', 'suborder_id'];
const columConfig: Record<string, PropertyColumnConfig> = {
  created_at: { sort: true },
  match_applied_core: { sort: true, align: 'right' },
  suborder_id: {
    render: ({ cell, data }) =>
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            const searchQs = useSearchQs();
            routerAction.open({
              name: MENU_BUSINESS_TICKET_MANAGEMENT,
              query: {
                [GLOBAL_BIZS_KEY]: data.bk_biz_id,
                type: 'host_recycle',
                filter: searchQs.build({ suborder_id: [cell], bk_username: [] }),
              },
            });
          },
        },
        cell,
      ),
  }, // 主机回收的子单据ID
};
const columns: ModelPropertyColumn[] = fieldIds.map((id) => ({
  ...usageOrderViewProperties.find((view) => view.id === id),
  ...columConfig[id],
}));

defineExpose({ show });
</script>

<template>
  <bk-dialog v-model:is-show="isShow" title="退还记录" dialog-type="show" width="640px">
    <div class="flex-row base-info">
      <div class="item">
        <div class="label">业务：</div>
        <display-value :property="{ type: 'business' }" :value="data?.bk_biz_id" />
      </div>
      <div class="item">
        <div class="label">子单据ID：</div>
        <!-- 主机申请的子单据ID -->
        <display-value :property="{ type: 'string' }" :value="data?.suborder_id" />
      </div>
    </div>
    <bk-table row-hover="auto" :data="data?.returned_records" row-key="id" pagination>
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :sort="column.sort"
        :align="column.align"
        :render="column.render"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
    </bk-table>
  </bk-dialog>
</template>

<style scoped lang="scss">
.base-info {
  gap: 90px;

  .item {
    display: inline-flex;
    align-items: center;
    line-height: 40px;

    .label {
      color: #4d4f56;
    }
  }
}
</style>
