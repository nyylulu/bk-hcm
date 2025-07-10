<template>
  <bk-collapse class="collapse-panel-grid">
    <bk-collapse-panel
      :model-value="isCollapse"
      :title="title"
      icon="right-shape"
      alone
      @update:model-value="handleToggle"
    >
      <template #content>
        <grid-display :column="column" :fields="fields" :details="details" />
      </template>
    </bk-collapse-panel>
  </bk-collapse>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { ModelPropertyDisplay } from '@/model/typings';

import GridDisplay from './grid-display.vue';

interface IProps {
  title: string;
  column?: number;
  fields: ModelPropertyDisplay[];
  details: any;
}

withDefaults(defineProps<IProps>(), {
  column: 2,
});

const isCollapse = ref(false);
const handleToggle = (c: boolean) => (isCollapse.value = c);
</script>

<style lang="scss" scoped>
.collapse-panel-grid {
  :deep(.bk-collapse-title) {
    color: #313238;
    font-weight: 700;
  }

  :deep(.bk-collapse-content) {
    padding: 16px 24px;
    background: #fff;
  }
}
</style>
