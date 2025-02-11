<script setup lang="ts"></script>

<template>
  <div class="i-timeline-content">
    <slot></slot>
  </div>
</template>

<style scoped lang="scss">
.i-timeline-content {
  font-size: 12px;
  line-height: 20px;
  color: #4d4f56;

  &.time-value,
  .time-value {
    color: #979ba5;
  }
}
</style>
