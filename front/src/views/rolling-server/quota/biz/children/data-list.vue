<script setup lang="ts">
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import quotaBizViewProperties from '@/model/rolling-server/quota-biz.view';
import type { IBizViewDataListProps } from '../../typings';
import type { IRollingServerBizQuotaItem } from '@/store/rolling-server-quota';

withDefaults(defineProps<IBizViewDataListProps>(), {});

const emit = defineEmits<{
  adjust: [row: IRollingServerBizQuotaItem];
  'view-record': [row: IRollingServerBizQuotaItem];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const columnConfig: Record<string, PropertyColumnConfig> = {
  bk_biz_name: {},
  base_quota: { sort: true },
  adjust_type: {},
  quota_offset: {},
  quota_offset_final: {},
  updated_at: {},
  reviser: {},
};

const columns: ModelPropertyColumn[] = [];
for (const [fieldId, config] of Object.entries(columnConfig)) {
  columns.push({
    ...quotaBizViewProperties.find((prop) => prop.id === fieldId),
    ...config,
  });
}

const { settings } = useTableSettings(columns);
</script>

<template>
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
      :label="column.name"
      :sort="column.sort"
    >
      <template #default="{ row }">
        <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
      </template>
    </bk-table-column>
    <bk-table-column :label="'操作'" :show-overflow-tooltip="false">
      <template #default="{ row }">
        <div class="actions">
          <bk-button theme="primary" text @click="emit('adjust', row)">调整额度</bk-button>
          <bk-button theme="primary" text @click="emit('view-record', row)">操作记录</bk-button>
        </div>
      </template>
    </bk-table-column>
  </bk-table>
</template>
<style lang="scss" scoped>
.actions {
  display: flex;
  gap: 12px;
}
</style>
