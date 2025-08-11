<script setup lang="ts">
import { computed } from 'vue';
import { VendorEnum } from '@/common/constant';
import useTableSelection from '@/hooks/use-table-selection';
import { ModelPropertyColumn } from '@/model/typings';
import { getPrivateIPs } from '@/utils';

import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

interface IProps {
  suborders: any[];
}
const props = defineProps<IProps>();

const upgradeCvmList = computed(() => props.suborders.flatMap((item) => item.upgrade_cvm_list));

const columns: ModelPropertyColumn[] = [
  { id: 'instance_id', name: '主机ID', type: 'string' },
  { id: 'private_ip', name: '内网IP', type: 'string', render: ({ row }: any) => getPrivateIPs(row) },
  { id: 'bk_asset_id', name: '固资号', type: 'string' },
  { id: 'region_id', name: '地域', type: 'region' },
  { id: 'zone_id', name: '可用区', type: 'string' },
  { id: 'device_type', name: '原机型', type: 'string' },
  { id: 'target_instance_type', name: '目标机型', type: 'string' },
];

const getDisplayCompProps = (column: ModelPropertyColumn) => {
  const { id } = column;
  if (id === 'region_id') {
    return { vendor: VendorEnum.ZIYAN }; // 单据目前只有自研云支持
  }
  return {};
};

const { selections, handleSelectChange, handleSelectAll } = useTableSelection({
  isRowSelectable: () => true,
  rowKey: 'instance_id',
});
const ipContent = computed(() => selections.value.map((item: any) => getPrivateIPs(item)).join('\n'));
const assetIdContent = computed(() => selections.value.map((item: any) => item.bk_asset_id).join('\n'));
</script>

<template>
  <div class="upgrade-cvm-table-container">
    <div class="copy-wrap">
      <copy-to-clipboard :disabled="selections.length === 0" :content="ipContent" class="mb12">
        <bk-button theme="primary" :disabled="selections.length === 0">复制已勾选IP</bk-button>
      </copy-to-clipboard>
      <copy-to-clipboard :disabled="selections.length === 0" :content="assetIdContent" class="mb12">
        <bk-button theme="primary" :disabled="selections.length === 0">复制已勾选固资号</bk-button>
      </copy-to-clipboard>
    </div>
    <p class="mb12">机型配置调整记录</p>
    <bk-table
      row-hover="auto"
      row-key="instance_id"
      :data="upgradeCvmList"
      show-overflow-tooltip
      @select="handleSelectChange"
      @select-all="handleSelectAll"
    >
      <bk-table-column :width="30" :min-width="30" type="selection" />
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :render="column.render"
      >
        <template #default="{ row }">
          <display-value
            :property="column"
            :value="row[column.id]"
            :display="column?.meta?.display"
            v-bind="getDisplayCompProps(column)"
          />
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<style scoped lang="scss">
.upgrade-cvm-table-container {
  .copy-wrap {
    display: flex;
    align-items: center;
    gap: 12px;
  }
}
</style>
