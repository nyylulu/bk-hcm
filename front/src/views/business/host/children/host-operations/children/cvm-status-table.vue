<script setup lang="ts">
import type { ICvmListOperateStatus } from '@/store/cvm-operate';
import type { ModelPropertyColumn } from '@/model/typings';

defineOptions({ name: 'cvm-status-table' });
defineProps<{ list: ICvmListOperateStatus[]; columns: ModelPropertyColumn[]; hasDeleteCell?: boolean }>();
const emit = defineEmits<{
  delete: [number];
}>();

const handleDelete = (index: number) => {
  emit('delete', index);
};
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
      :width="column.width"
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
    <bk-table-column v-if="hasDeleteCell" label="操作" width="80" fixed="right">
      <template #default="{ index }">
        <bk-button text @click="handleDelete(index)">
          <i class="hcm-icon bkhcm-icon-minus-circle-shape delete-icon"></i>
        </bk-button>
      </template>
    </bk-table-column>
  </bk-table>
</template>

<style scoped lang="scss">
.delete-icon {
  width: 14px;
  height: 14px;
  color: #c4c6cc;
}
</style>
