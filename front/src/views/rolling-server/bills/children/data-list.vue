<script setup lang="ts">
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import billsViewProperties from '@/model/rolling-server/bills.view';
import { getColumnName } from '@/model/utils';
import type { IBillsDataListProps } from '../typings';
import type { IRollingServerBillItem } from '@/store/rolling-server-bills';

withDefaults(defineProps<IBillsDataListProps>(), {});

const emit = defineEmits<{
  'view-fine-details': [row: IRollingServerBillItem];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const columnConfig: Record<string, PropertyColumnConfig> = {
  bk_biz_id: {},
  date: { sort: true },
  not_returned_core: { sort: true, align: 'right' },
};

const columns: ModelPropertyColumn[] = [];
for (const [fieldId, config] of Object.entries(columnConfig)) {
  columns.push({
    ...billsViewProperties.find((prop) => prop.id === fieldId),
    ...config,
  });
}

const { settings } = useTableSettings(columns);

const getDisplayCompProps = (column: ModelPropertyColumn) => {
  const { id } = column;
  if (id === 'date') {
    return { format: 'YYYY-MM-DD' };
  }
  return {};
};
</script>

<template>
  <div class="bills-data-list">
    <bk-table
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
    >
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="getColumnName(column)"
        :sort="column.sort"
        :align="column.align"
      >
        <template #default="{ row }">
          <display-value
            :property="column"
            :value="row[column.id]"
            :display="column?.meta?.display"
            v-bind="getDisplayCompProps(column)"
          />
        </template>
      </bk-table-column>
      <bk-table-column :label="'操作'" :show-overflow-tooltip="false">
        <template #default="{ row }">
          <bk-button theme="primary" text @click="emit('view-fine-details', row)">查看关联单据</bk-button>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>
<style lang="scss" scoped>
.bills-data-list {
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  padding: 16px 24px;
}
</style>
