<script setup lang="ts">
import type { ICvmListRestStatus } from '@/store/cvm/reset';

import columns from './columns';

const props = defineProps<{ list: ICvmListRestStatus[] }>();

const renderColumns = columns.slice();
</script>

<template>
  <bk-table
    class=""
    row-hover="auto"
    :data="props.list"
    min-height="auto"
    max-height="300px"
    show-overflow-tooltip
    row-key="id"
  >
    <bk-table-column
      v-for="(column, index) in renderColumns"
      :key="index"
      :prop="column.id"
      :label="column.name"
      :render="column.render"
    >
      <template #default="{ row }">
        <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
      </template>
    </bk-table-column>
  </bk-table>
</template>

<style scoped lang="scss"></style>
