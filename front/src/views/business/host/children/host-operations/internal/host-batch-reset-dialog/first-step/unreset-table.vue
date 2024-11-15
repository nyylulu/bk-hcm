<script setup lang="ts">
import { ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ICvmListRestStatus } from '@/store/cvm/reset';
import { RESET_STATUS_MAP } from '../constants';

import columns from './columns';

const props = defineProps<{ list: ICvmListRestStatus[] }>();

const { t } = useI18n();

const renderColumns = columns.slice(1);

const activeIndex = ref([0]);
const renderList = ref([]);
watchEffect(() => {
  const renderListMap = props.list.reduce((map, item) => {
    const key = item.reset_status;
    if (!map.has(key)) map.set(key, []);
    map.get(key).push(item);
    return map;
  }, new Map<number, ICvmListRestStatus[]>());

  renderList.value = Array.from(renderListMap, ([key, value]) => ({ key, name: RESET_STATUS_MAP[key], value }));
});
</script>

<template>
  <bk-collapse v-model="activeIndex" use-block-theme v-if="renderList.length">
    <bk-collapse-panel v-for="(item, index) in renderList" :key="index" :name="index">
      <span class="name">{{ item.name }}</span>
      <span class="count">{{ `${t(`（`)}${item.value.length}${t('）')}` }}</span>

      <template #content>
        <bk-table
          row-hover="auto"
          :data="item.value"
          min-height="auto"
          max-height="300px"
          show-overflow-tooltip
          row-key="id"
        >
          <bk-table-column
            v-for="(column, columnIndex) in renderColumns"
            :key="columnIndex"
            :prop="column.id"
            :label="column.name"
            :render="column.render"
          >
            <template #default="{ row }">
              <display-value
                :property="column"
                :value="row[column.id]"
                :display="column?.meta?.display"
                :vendor="row?.vendor"
              />
            </template>
          </bk-table-column>
        </bk-table>
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
.name {
  font-size: 12px;
  color: #313238;
}
.count {
  font-size: 12px;
  color: #979ba5;
}
</style>
