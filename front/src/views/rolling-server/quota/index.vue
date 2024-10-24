<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import bizView from './biz.vue';
import globalView from './global.vue';

const router = useRouter();
const route = useRoute();

const tabPanels = [
  { name: 'global', label: '资源池额度' },
  { name: 'biz', label: '业务额度' },
];
const viewActive = computed({
  get() {
    return (route.params.view as string) || 'global';
  },
  set(value) {
    router.push({ params: { view: value } });
  },
});

const viewComps: Record<string, any> = {
  biz: bizView,
  global: globalView,
};
</script>

<template>
  <bk-tab v-model:active="viewActive" type="card-grid" class="roll-server-quota">
    <bk-tab-panel v-for="panel in tabPanels" :key="panel.name" :label="panel.label" :name="panel.name">
      <component :is="viewComps[viewActive]" />
    </bk-tab-panel>
  </bk-tab>
</template>

<style lang="scss" scoped>
.roll-server-quota {
  height: 100%;
  :deep(.bk-tab-header) {
    // 嵌套tab组件，受上层tab组件影响，需要重置
    padding: 0 !important;
    background: none !important;
    box-shadow: none !important;
  }
}
</style>
