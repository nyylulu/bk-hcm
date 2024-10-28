<script setup lang="ts">
import { ref, watchEffect } from 'vue';
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import { useRollingServerQuotaStore, type IRollingServerBizQuotaItem } from '@/store/rolling-server-quota';
import quotaBizViewProperties from '@/model/rolling-server/quota-biz.view';

const props = defineProps<{ dataRow: IRollingServerBizQuotaItem }>();
const model = defineModel<boolean>();

const rollingServerQuotaStore = useRollingServerQuotaStore();

const list = ref([]);

const columnConfig: Record<string, PropertyColumnConfig> = {
  created_at: {},
  adjust_type: {},
  quota_offset: {},
  reviser: {},
};

const columns: ModelPropertyColumn[] = [];
for (const [fieldId, config] of Object.entries(columnConfig)) {
  columns.push({
    ...quotaBizViewProperties.find((prop) => prop.id === fieldId),
    ...config,
  });
}

watchEffect(async () => {
  list.value = await rollingServerQuotaStore.getAdjustRecords({ offset_config_ids: [props.dataRow.offset_config_id] });
});
</script>

<template>
  <bk-dialog dialog-type="show" title="操作记录" width="640" v-model:is-show="model">
    <bk-table row-hover="auto" :data="list" show-overflow-tooltip>
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
