<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { ISlaSetItem, useLoadBalancerStore } from '@/store';
import { ModelPropertyColumn } from '@/model/typings';
import bus from '@/common/bus';
import dialogFooter from '@/components/common-dialog/dialog-footer.vue';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { CLB_SPECS } from '@/common/constant';
import { getSimpleConditionBySearchSelect } from '@/utils/search';

interface IProps {
  accountId: string;
  region: string;
  slaTypes?: string[];
  slaType: string;
}

const props = defineProps<IProps>();
const emit = defineEmits<{
  confirm: [slaType: string];
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
  },
  {
    id: 'MaxConn',
    name: '并发连接数上限(个/秒)',
    type: 'number',
  },
  {
    id: 'MaxCps',
    name: '新建连接数上限(个/秒)',
    type: 'number',
  },
  {
    id: 'MaxOutBits',
    name: '最大出口流量(Mbps)',
    type: 'number',
  },
  {
    id: 'MaxInBits',
    name: '最大入口流量(Mbps)',
    type: 'number',
  },
  {
    id: 'MaxQps',
    name: '最大pqs(个/秒)',
    type: 'number',
  },
];

const isShow = ref(false);
const datalist = ref<ISlaSetItem[]>([]);
const selected = ref<string>(props.slaType);

const getSlaCapacityDescribe = async (slaTypes?: string[]) => {
  const { accountId, region, slaType } = props;
  if (accountId && region) {
    const list = await loadbalancerStore.queryZiyanLoadBalancerSlaCapacityDescribe({
      account_id: accountId,
      region,
      sla_types: slaTypes,
    });
    datalist.value = list.filter((item) => item.SlaType !== 'clb.c1.small');

    // 回显
    selected.value = slaType;
  }
};
watch(
  isShow,
  async (val) => {
    if (!val) return;
    getSlaCapacityDescribe();
  },
  { immediate: true },
);

const handleRowClick = (_event: PointerEvent, row: ISlaSetItem) => {
  selected.value = row.SlaType;
};

const searchValue = ref([]);
const searchData: ISearchItem[] = [
  {
    id: 'sla_types',
    name: '性能容量型规格',
    children: Object.entries(CLB_SPECS).map(([id, name]) => ({ id, name })),
    multiple: true,
  },
];
watch(searchValue, (val) => {
  const { sla_types } = getSimpleConditionBySearchSelect(val);
  getSlaCapacityDescribe(sla_types);
});

const handleConfirm = () => {
  emit('confirm', selected.value);
  nextTick(() => {
    handleClosed();
  });
};
const handleClosed = () => {
  isShow.value = false;
  selected.value = undefined;
  searchValue.value = [];
};

onMounted(() => {
  bus.$on('showZiyanLbSpecTypeSelectDialog', () => {
    isShow.value = true;
  });
});
onUnmounted(() => {
  bus.$off('showZiyanLbSpecTypeSelectDialog');
});
</script>

<template>
  <bk-dialog v-model:is-show="isShow" title="选择实例规格" width="60vw" @closed="handleClosed">
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
      <bk-table-column v-for="(column, index) in columns" :key="index" :prop="column.id" :label="column.name">
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
