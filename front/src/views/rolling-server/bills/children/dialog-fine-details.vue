<script setup lang="ts">
import { ref, watchEffect } from 'vue';
import type { ModelPropertyColumn, PropertyColumnConfig } from '@/model/typings';
import { useRollingServerBillsStore, type IRollingServerBillItem } from '@/store/rolling-server-bills';
import usePage from '@/hooks/use-page';
import billsDetailsViewProperties from '@/model/rolling-server/bills-details.view';
import { transformSimpleCondition } from '@/utils/search';
import { getColumnName } from '@/model/utils';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import BusinessValue from '@/components/display-value/business-value.vue';

const props = defineProps<{ dataRow: IRollingServerBillItem }>();
const model = defineModel<boolean>();

const rollingServerBillsStore = useRollingServerBillsStore();
const { pagination, getPageParams, handlePageChange, handlePageSizeChange, handleSort } = usePage(false);

const detailsList = ref([]);

const columnConfig: Record<string, PropertyColumnConfig> = {
  suborder_id: {},
  delivered_core: {},
  returned_core: {},
  not_returned_core: {},
};

const columns: ModelPropertyColumn[] = [];
for (const [fieldId, config] of Object.entries(columnConfig)) {
  columns.push({
    ...billsDetailsViewProperties.find((prop) => prop.id === fieldId),
    ...config,
  });
}

watchEffect(async () => {
  const condition = {
    bk_biz_id: props.dataRow.bk_biz_id,
    year: props.dataRow.year,
    month: props.dataRow.month,
    day: props.dataRow.day,
  };
  const { list, count } = await rollingServerBillsStore.getBillFineDetailsList({
    filter: transformSimpleCondition(condition, billsDetailsViewProperties),
    page: getPageParams(pagination),
  });

  detailsList.value = list;
  // detailsList.value = [
  //   {
  //     id: 'aa1111',
  //     bk_biz_id: 3232,
  //     order_id: 'aaaa1111',
  //     suborder_id: 'bbb1111',
  //     year: 2024,
  //     month: 10,
  //     day: 29,
  //     delivered_core: 2323,
  //     returned_core: 32,
  //     creator: 'test',
  //     created_at: '2024-10-02T15:04:05Z',
  //   },
  // ];
  pagination.count = count || 100;
});
</script>

<template>
  <bk-dialog dialog-type="show" title="关联单据详情" width="960" v-model:is-show="model">
    <grid-container :column="2" :fixed="true" :gap="36" label-align="left" label-width="auto" content-min-width="auto">
      <grid-item label="业务"><business-value :value="dataRow.bk_biz_id" /></grid-item>
      <grid-item label="核算日期">{{ `${dataRow.year}-${dataRow.month}-${dataRow.day}` }}</grid-item>
    </grid-container>
    <bk-table
      row-hover="auto"
      remote-pagination
      show-overflow-tooltip
      :data="detailsList"
      :pagination="pagination"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    >
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="getColumnName(column)"
        :sort="column.sort"
        :align="column.align"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
    </bk-table>
  </bk-dialog>
</template>
