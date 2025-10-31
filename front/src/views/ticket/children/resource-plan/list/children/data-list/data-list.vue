<script setup lang="ts">
import { inject } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import { type IResourcePlanTicketItem } from '@/store/ticket/resource-plan';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: IResourcePlanTicketItem[];
  pagination: PaginationType;
}

const props = withDefaults(defineProps<IDataListProps>(), {});

const emit = defineEmits<{
  'view-details': [row: IResourcePlanTicketItem];
}>();

const isServicePage = inject('isServicePage');

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);
</script>

<template>
  <bk-table
    row-hover="auto"
    :data="list"
    :pagination="pagination"
    :max-height="`calc(100vh - ${isServicePage ? 452 : 452}px)`"
    :settings="settings"
    remote-pagination
    show-overflow-tooltip
    @page-limit-change="handlePageSizeChange"
    @page-value-change="handlePageChange"
    @column-sort="handleSort"
    row-key="id"
  >
    <bk-table-column :label="'预测单号'">
      <template #default="{ row }: { row: IResourcePlanTicketItem }">
        <bk-button theme="primary" text @click="emit('view-details', row)">{{ row.id }}</bk-button>
      </template>
    </bk-table-column>
    <bk-table-column
      v-for="(column, index) in columns"
      :key="index"
      :prop="column.id"
      :label="column.name"
      :sort="column.sort"
      :align="column.align"
      :width="column.width"
      :render="column.render"
    >
      <template #default="{ row }">
        <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
      </template>
    </bk-table-column>
  </bk-table>
</template>
