<template>
  <grid-item v-bind="props">
    <div class="form-element">
      <slot />
    </div>
  </grid-item>
</template>
<script setup lang="ts">
import GridItem, { type IGridItemProps } from './grid-item.vue';
const props = defineProps<IGridItemProps>();
</script>
