<script setup lang="ts">
import { ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ICvmListOperateStatus } from '@/store/cvm-operate';
import { OPERATE_STATUS_MAP } from '../constants';

import cvmStatusBaseColumns from '../constants/cvm-status-base-columns';
import cvmStatusTable from './cvm-status-table.vue';

defineOptions({ name: 'cvm-status-collapse-table' });
const props = defineProps<{ list: ICvmListOperateStatus[] }>();

const { t } = useI18n();

const renderColumns = cvmStatusBaseColumns.slice(1);

const activeIndex = ref([0]);
const renderList = ref([]);
watchEffect(() => {
  const renderListMap = props.list.reduce((map, item) => {
    const key = item.operate_status;
    if (!map.has(key)) map.set(key, []);
    map.get(key).push(item);
    return map;
  }, new Map<number, ICvmListOperateStatus[]>());

  renderList.value = Array.from(renderListMap, ([key, value]) => ({ key, name: OPERATE_STATUS_MAP[key], value }));
});
</script>

<template>
  <bk-collapse v-if="renderList.length" v-model="activeIndex" use-block-theme class="cvm-status-collapse-table">
    <bk-collapse-panel v-for="(item, index) in renderList" :key="index" :name="index" icon="right-shape">
      <span class="name">{{ item.name }}</span>
      <span class="count">{{ `${t(`（`)}${item.value.length}${t('）')}` }}</span>

      <template #content>
        <cvm-status-table :list="item.value" :columns="renderColumns" />
      </template>
    </bk-collapse-panel>
  </bk-collapse>
  <bk-exception
    v-else
    class="exception-wrap-item exception-part"
    :description="t('暂无数据')"
    scene="part"
    type="empty"
  />
</template>

<style scoped lang="scss">
.cvm-status-collapse-table {
  :deep(.bk-collapse-header) {
    background: #ffebeb;

    .bk-collapse-icon {
      width: 12px;
    }
  }

  .name {
    color: #313238;
  }

  .count {
    color: #979ba5;
  }
}
</style>
