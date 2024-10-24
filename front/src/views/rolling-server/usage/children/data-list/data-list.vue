<script setup lang="ts">
import type { IDataListProps } from '../../typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import columnFactory from './column-factory';

const { getColumns } = columnFactory();

const props = withDefaults(defineProps<IDataListProps>(), {});

const emit = defineEmits<{
  'view-details': [id: string];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();
const columns = getColumns(props.view);
console.error('🚀 ~ columns:', columns);

const { settings } = useTableSettings(columns);
</script>

<template>
  <div class="rolling-server-usage-data-list">
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
          <bk-button theme="primary" text @click="emit('view-details', row.id)">查看详情</bk-button>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<style scoped></style>
