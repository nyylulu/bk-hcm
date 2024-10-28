<script setup lang="ts">
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import type { IDataListProps } from '../../typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import usageAppliedView from '@/model/rolling-server/usage-applied.view';
import usageReturnedView from '@/model/rolling-server/usage-returned.view';

withDefaults(defineProps<IDataListProps>(), {});

const emit = defineEmits<{
  'view-details': [id: string];
  'show-returned-records': [id: string];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const usageView = [...usageAppliedView, ...usageReturnedView];
const fieldIds = [
  'order_id',
  'bk_biz_id',
  'created_at',
  'applied_type',
  'applied_core',
  'delivered_core',
  'returned_core',
  'not_returned_core',
  'exec_rate',
  'creator',
];
const columConfig: Record<string, PropertyColumnConfig> = {
  applied_core: { sort: true },
  delivered_core: { sort: true },
};
const columns: ModelPropertyColumn[] = fieldIds.map((id) => ({
  ...usageView.find((view) => view.id === id),
  ...columConfig[id],
}));

const { settings } = useTableSettings(columns);
</script>

<template>
  <div class="rolling-server-usage-data-list">
    <div class="table-tools">
      <bk-button>导出</bk-button>
      <ul class="summary">
        <li class="item">
          <div class="label">总交付（CPU核数）：</div>
          <div class="value">150,000</div>
        </li>
        <li class="item">
          <div class="label">总返还（CPU核数）：</div>
          <div class="value">140,000</div>
        </li>
      </ul>
    </div>
    <bk-table
      ref="tableRef"
      row-hover="auto"
      :data="list"
      :pagination="pagination"
      :max-height="'calc(100vh - 401px)'"
      :settings="settings"
      remote-pagination
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      row-key="id"
    >
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :sort="column.sort"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
      <bk-table-column :label="'操作'">
        <template #default="{ row }">
          <bk-button theme="primary" text @click="emit('show-returned-records', row.id)">返还记录</bk-button>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<style scoped lang="scss">
.rolling-server-usage-data-list {
  padding: 16px 24px;
  background-color: #fff;

  .table-tools {
    margin-bottom: 12px;
    display: flex;
    align-items: center;

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
}
</style>
