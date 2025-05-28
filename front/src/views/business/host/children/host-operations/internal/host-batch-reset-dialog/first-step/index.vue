<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRegionsStore } from '@/store/useRegionsStore';
import { getLocalFilterFnBySearchSelect } from '@/utils/search';
import type { CvmListRestDataView } from '../../../typings';

import cvmStatusBaseColumns from '../../../constants/cvm-status-base-columns';

import cvmStatusTable from '../../../children/cvm-status-table.vue';
import cvmStatusCollapseTable from '../../../children/cvm-status-collapse-table.vue';
import { ICvmListOperateStatus } from '@/store/cvm-operate';

const props = defineProps<{
  listData: CvmListRestDataView;
  nonIdleCvmList: ICvmListOperateStatus[];
  loading: boolean;
}>();
const emit = defineEmits<{
  delete: [number];
}>();

const { t } = useI18n();
const regionsStore = useRegionsStore();

const radios = [
  { label: 0, alias: t('可重装') },
  { label: 1, alias: t('不可重装') },
];
const selected = ref(radios[1].label);
const isResettable = computed(() => selected.value === radios[0].label);
watch(
  () => props.listData.reset,
  (list) => {
    selected.value = list.length > 0 ? radios[0].label : radios[1].label;
  },
  { immediate: true, deep: true },
);

const searchData = computed(() => [
  { id: 'private_ip_address', name: t('内网IP') },
  { id: 'bk_host_name', name: t('实例名称') },
]);
const searchValue = ref();

const renderList = computed(() => {
  const { reset, unReset } = props.listData;

  // 根据search-select过滤
  const filterFn = getLocalFilterFnBySearchSelect(searchValue.value, [
    { field: 'region', formatter: (v: string) => regionsStore.getRegionNameEN(v) },
  ]);

  return isResettable.value ? reset.filter(filterFn) : unReset.filter(filterFn);
});

const handleDelete = (index: number) => {
  emit('delete', index);
};
</script>

<template>
  <div class="i-first-step-container">
    <section class="i-tools-wrap">
      <bk-radio-group v-model="selected">
        <bk-radio-button v-for="{ label, alias } in radios" :key="label" :label="label" class="button">
          {{ alias }}
        </bk-radio-button>
      </bk-radio-group>
      <div class="info">
        {{ t('已选择') }}
        <span class="count text-primary">{{ listData.count }}</span>
        {{ t('个主机，其中可重装') }}
        <span class="count text-success">{{ listData.reset.length }}</span>
        {{ t('个，不可重装') }}
        <span class="count text-danger">{{ listData.unReset.length }}</span>
        {{ t('个。') }}
      </div>
      <!-- 本地搜索 -->
      <bk-search-select v-model="searchValue" :data="searchData" value-behavior="need-key" class="search" />
    </section>

    <div v-bkloading="{ loading }" style="min-height: 200px">
      <!-- loading的z-index小于表格表头 -->
      <template v-if="!loading">
        <!-- 可重装 -->
        <template v-if="isResettable">
          <bk-alert v-if="nonIdleCvmList.length > 0" theme="danger" class="mt16 mb16">
            <template #title>
              {{ t('有') }}
              {{ nonIdleCvmList.length }}
              {{ t('个主机处于非空闲机模块，请确认所选主机是否正确') }}
            </template>
          </bk-alert>
          <cvm-status-table
            v-if="isResettable"
            :list="renderList"
            :columns="cvmStatusBaseColumns"
            has-delete-cell
            @delete="handleDelete"
          />
        </template>
        <!-- 不可重装 -->
        <cvm-status-collapse-table v-else :list="renderList" />
      </template>
    </div>
  </div>
</template>

<style scoped lang="scss">
.i-first-step-container {
  .i-tools-wrap {
    margin-bottom: 12px;
    display: flex;
    align-items: center;
    .button {
      min-width: 88px;
    }
    .info {
      margin-left: 12px;
      line-height: 22px;
      color: #313238;
      .count {
        margin: 0 3px;
        font-weight: 700;
      }
    }
    .search {
      margin-left: auto;
      width: 400px;
    }
  }
}
</style>
