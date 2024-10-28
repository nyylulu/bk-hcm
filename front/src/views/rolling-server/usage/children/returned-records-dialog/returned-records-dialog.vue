<script setup lang="ts">
import { ref } from 'vue';
import { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import usageReturnedView from '@/model/rolling-server/usage-returned.view';
import { RollingServerRecordItem } from '@/store';

const isShow = ref(false);
const data = ref<RollingServerRecordItem>();
const show = (row: RollingServerRecordItem) => {
  isShow.value = true;
  data.value = row;
};
const fieldIds = ['created_at', 'match_applied_core', 'applied_record_id'];
const columConfig: Record<string, PropertyColumnConfig> = {
  match_applied_core: { sort: true },
};
const columns: ModelPropertyColumn[] = fieldIds.map((id) => ({
  ...usageReturnedView.find((view) => view.id === id),
  ...columConfig[id],
}));

defineExpose({ show });
</script>

<template>
  <bk-dialog v-model:is-show="isShow" title="退还记录" dialog-type="show" width="640px">
    <div class="flex-row base-info">
      <div class="item">
        <div class="label">业务：</div>
        <div>和平精英</div>
      </div>
      <div class="item">
        <div class="label">单据ID：</div>
        <div>REQ20240120000001</div>
      </div>
    </div>
    <bk-table row-hover="auto" :data="data?.returned_records" row-key="id" pagination>
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
