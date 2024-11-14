<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { ResourceTypeEnum } from '@/common/resource-constant';
import clb from './clb.vue';
import cvm from './cvm.vue';

const route = useRoute();

const resourceType = computed(() => (route.params.resourceType || ResourceTypeEnum.CLB) as string);

const detailComps: Record<string, any> = { clb, cvm };
</script>

<template>
  <div class="page-task-details">
    <component :is="detailComps[resourceType]" />
  </div>
</template>

<style lang="scss" scoped>
.page-task-details {
  padding: 24px;
  height: 100%;
}
</style>
