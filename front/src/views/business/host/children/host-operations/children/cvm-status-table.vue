<script setup lang="ts">
import type { ICvmListOperateStatus } from '@/store/cvm-operate';
import type { ModelPropertyColumn } from '@/model/typings';

defineOptions({ name: 'cvm-status-table' });
defineProps<{ list: ICvmListOperateStatus[]; columns: ModelPropertyColumn[] }>();
</script>

<template>
  <bk-table
    class=""
    row-hover="auto"
    :data="list"
    min-height="auto"
    max-height="300px"
    show-overflow-tooltip
    row-key="id"
  >
    <bk-table-column
      v-for="(column, index) in columns"
      :key="index"
      :prop="column.id"
      :label="column.name"
      :render="column.render"
    >
      <template #default="{ row }">
        <display-value
          :property="column"
          :value="row[column.id]"
          :display="column?.meta?.display"
          :vendor="row?.vendor"
        />
      </template>
    </bk-table-column>
  </bk-table>
</template>

<style scoped lang="scss"></style>
