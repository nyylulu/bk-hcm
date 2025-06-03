<script setup lang="ts">
import { ref, watch, watchEffect } from 'vue';
import { ISlaSetItem, useLoadBalancerStore } from '@/store';
import { getSimpleConditionBySearchSelect } from '@/utils/search';
import { CLB_SPECS } from '@/common/constant';
import { ModelPropertyColumn } from '@/model/typings';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';

import dialogFooter from '@/components/common-dialog/dialog-footer.vue';

interface IProps {
  accountId: string;
  region: string;
  slaTypes?: string[];
  slaType: string;
}

const model = defineModel<boolean>();
const props = defineProps<IProps>();
const emit = defineEmits<{
  confirm: [{ slaType: string; bandwidthLimit: number }];
  hidden: [];
}>();

const loadbalancerStore = useLoadBalancerStore();

const columns: ModelPropertyColumn[] = [
  {
    id: 'SlaType',
    name: '性能容量型规格',
    type: 'string',
  },
  {
    id: 'SlaName',
    name: '规格名称',
    type: 'string',
    width: 100,
  },
  {
    id: 'MaxConn',
    name: '并发连接数上限(个/秒)',
    type: 'number',
    meta: { display: { format: (value) => value?.toLocaleString('en-US') } },
    width: 180,
  },
  {
    id: 'MaxCps',
    name: '新建连接数上限(个/秒)',
    type: 'number',
    meta: { display: { format: (value) => value?.toLocaleString('en-US') } },
    width: 180,
  },
  {
    id: 'MaxOutBits',
    name: '最大出口流量(Mbps)',
    type: 'number',
    meta: { display: { format: (value) => value?.toLocaleString('en-US') } },
  },
  {
    id: 'MaxInBits',
    name: '最大入口流量(Mbps)',
    type: 'number',
    meta: { display: { format: (value) => value?.toLocaleString('en-US') } },
  },
  {
    id: 'MaxQps',
    name: '最大pqs(个/秒)',
    type: 'number',
    meta: { display: { format: (value) => value?.toLocaleString('en-US') } },
  },
];
const selected = ref<string>(props.slaType);
const datalist = ref<ISlaSetItem[]>([]);

const getSlaCapacityDescribe = async (slaTypes?: string[]) => {
  const { accountId, region } = props;
  if (accountId && region) {
    const list = await loadbalancerStore.queryZiyanLoadBalancerSlaCapacityDescribe({
      account_id: accountId,
      region,
      sla_types: slaTypes,
    });
    datalist.value = list.filter((item) => item.SlaType !== 'clb.c1.small');
  }
};
watchEffect(() => {
  if (model.value) {
    getSlaCapacityDescribe();
  }
});

const handleRowClick = (_event: PointerEvent, row: ISlaSetItem) => {
  selected.value = row.SlaType;
};

const searchValue = ref([]);
const searchData: ISearchItem[] = [
  {
    id: 'sla_types',
    name: '性能容量型规格',
    children: Object.entries(CLB_SPECS)
      .map(([id, name]) => ({ id, name }))
      .filter((item) => item.id !== 'clb.c1.small'),
    multiple: true,
  },
];
watch(searchValue, (val) => {
  const { sla_types } = getSimpleConditionBySearchSelect(val);
  getSlaCapacityDescribe(sla_types);
});

const handleConfirm = () => {
  const result = datalist.value.find((item) => item.SlaType === selected.value);
  emit('confirm', { slaType: result.SlaType, bandwidthLimit: result.MaxOutBits });
  handleClosed();
};
const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" title="选择实例规格" width="60vw" @closed="handleClosed" @hidden="emit('hidden')">
    <bk-search-select v-model="searchValue" :data="searchData" class="mb12" />
    <bk-table
      :data="datalist"
      row-hover="auto"
      show-overflow-tooltip
      v-bkloading="{ loading: loadbalancerStore.queryZiyanLoadBalancerSlaCapacityDescribeLoading }"
      @row-click="handleRowClick"
    >
      <bk-table-column width="40" min-width="40" :show-overflow-tooltip="false">
        <template #default="{ row }">
          <bk-radio v-model="selected" :label="row.SlaType">　</bk-radio>
        </template>
      </bk-table-column>
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :width="column.width"
      >
        <template #default="{ row }">
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </bk-table-column>
    </bk-table>
    <template #footer>
      <dialog-footer :disabled="!selected" @confirm="handleConfirm" @closed="handleClosed" />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss"></style>
