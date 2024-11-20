<script setup lang="ts">
import usePage from '@/hooks/use-page';
import { getModel } from '@/model/manager';
import { getColumnName } from '@/model/utils';
import { ModelPropertyColumn } from '@/model/typings';
import { StatsListView } from '@/model/green-channel/stats-list.view';
import type { IDataListProps } from '../typings';

withDefaults(defineProps<IDataListProps>(), {});

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const properties = getModel(StatsListView).getProperties<ModelPropertyColumn>();
</script>

<template>
  <div class="stats-list">
    <bk-table
      row-hover="auto"
      :data="list"
      :pagination="pagination"
      :max-height="'calc(100vh - 401px)'"
      remote-pagination
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    >
      <bk-table-column
        v-for="(column, index) in properties"
        :key="index"
        :prop="column.id"
        :label="getColumnName(column)"
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
  </div>
</template>

<style lang="scss" scoped>
.stats-list {
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  padding: 16px 24px;
}
</style>
