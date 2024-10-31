<script setup lang="ts">
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import { getTableNewRowClass } from '@/common/util';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import quotaBizViewProperties from '@/model/rolling-server/quota-biz.view';
import type { IRollingServerBizQuotaItem } from '@/store/rolling-server-quota';
import { QuotaAdjustType } from '@/views/rolling-server/typings';
import type { IBizViewDataListProps } from '../../typings';

withDefaults(defineProps<IBizViewDataListProps>(), {});

const emit = defineEmits<{
  adjust: [row: IRollingServerBizQuotaItem];
  'view-record': [row: IRollingServerBizQuotaItem];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const columnConfig: Record<string, PropertyColumnConfig> = {
  bk_biz_name: {},
  quota: {
    width: 120,
    align: 'right',
    sort: true,
    render: ({ cell }: { cell?: IRollingServerBizQuotaItem['quota'] }) => cell ?? '--',
  },
  adjust_type: { width: 220 },
  quota_offset: {
    width: 120,
    align: 'right',
    render: ({ data }: { data?: IRollingServerBizQuotaItem }) => {
      if (data.quota_offset) {
        return `${data.adjust_type === QuotaAdjustType.INCREASE ? '+' : '-'}${data.quota_offset}`;
      }
      return '--';
    },
  },
  quota_offset_final: {
    width: 200,
    align: 'right',
    render: ({ data }: { data?: IRollingServerBizQuotaItem }) => {
      if (data.quota) {
        return data.quota + (data.quota_offset ?? 0) * (data.adjust_type === QuotaAdjustType.INCREASE ? 1 : -1);
      }
      return '--';
    },
  },
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
    class="biz-quota-list"
    row-hover="auto"
    :data="list"
    :pagination="pagination"
    :max-height="'calc(100vh - 424px)'"
    :settings="settings"
    :row-class="getTableNewRowClass"
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
      :align="column.align"
      :width="column.width"
      :render="column.render"
    >
      <template #default="{ row }">
        <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
      </template>
    </bk-table-column>
    <bk-table-column :label="'操作'" :show-overflow-tooltip="false">
      <template #default="{ row }">
        <div class="actions">
          <bk-button theme="primary" text @click="emit('adjust', row)">调整额度</bk-button>
          <bk-button theme="primary" text :disabled="!row.offset_config_id" @click="emit('view-record', row)">
            操作记录
          </bk-button>
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
.biz-quota-list {
  :deep(.table-new-row) {
    td {
      background-color: #f2fff4 !important;
    }
  }
}
</style>
