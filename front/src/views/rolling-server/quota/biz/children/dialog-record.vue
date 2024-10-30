<script setup lang="ts">
import { ref, watchEffect } from 'vue';
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import { QuotaAdjustType } from '@/views/rolling-server/typings';
import { useRollingServerQuotaStore, type IRollingServerBizQuotaItem } from '@/store/rolling-server-quota';
import quotaBizRecordProperties from '@/model/rolling-server/quota-biz-record.view';

const props = defineProps<{ dataRow: IRollingServerBizQuotaItem }>();
const model = defineModel<boolean>();

const rollingServerQuotaStore = useRollingServerQuotaStore();

const list = ref([]);

const columnConfig: Record<string, PropertyColumnConfig> = {
  created_at: { width: 150 },
  adjust_type: {},
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
  operator: { width: 220 },
};

const columns: ModelPropertyColumn[] = [];
for (const [fieldId, config] of Object.entries(columnConfig)) {
  columns.push({
    ...quotaBizRecordProperties.find((prop) => prop.id === fieldId),
    ...config,
  });
}

watchEffect(async () => {
  list.value = await rollingServerQuotaStore.getAdjustRecords({ offset_config_ids: [props.dataRow.offset_config_id] });
});
</script>

<template>
  <bk-dialog dialog-type="show" title="操作记录" width="640" v-model:is-show="model">
    <bk-table row-hover="auto" :data="list" show-overflow-tooltip :max-height="500">
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
  </bk-dialog>
</template>
