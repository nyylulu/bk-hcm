<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';

import quota from './quota/index.vue';
import usage from './usage/index.vue';

const router = useRouter();
const route = useRoute();

const tabPanels = [
  { name: 'quota', label: '额度管理' },
  { name: 'usage', label: '额度执行' },
  { name: 'review', label: '滚服核算' },
];
const tabActive = computed({
  get() {
    return route.params.module || tabPanels[0].name;
  },
  set(value: string) {
    router.push({ params: { module: value, view: '' } });
  },
});

const tabComps: Record<string, any> = { quota, usage };
</script>

<template>
  <bk-tab class="page-roll-server" type="unborder-card" v-model:active="tabActive">
    <bk-tab-panel
      v-for="panel in tabPanels"
      :key="panel.name"
      :name="panel.name"
      :label="panel.label"
      render-directive="'if'"
    >
      <component :is="tabComps[tabActive]" v-if="tabActive === panel.name" />
    </bk-tab-panel>
  </bk-tab>
</template>

<style lang="scss" scoped>
.page-roll-server {
  height: 100%;
  :deep(.bk-tab-header) {
    padding: 0 12px;
    background: #fff;
    border-bottom: none;
    box-shadow: 0 3px 4px 0 #0000000a;
  }
  :deep(.bk-tab-content) {
    padding: 24px;
  }
}
</style>
