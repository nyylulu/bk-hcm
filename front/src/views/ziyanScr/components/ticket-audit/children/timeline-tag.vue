<script setup lang="ts">
interface IProps {
  isCurrent?: boolean;
}
withDefaults(defineProps<IProps>(), {
  isCurrent: false,
});
</script>

<template>
  <div class="i-timeline-tag" :class="[{ 'is-current': isCurrent }]">
    <slot></slot>
  </div>
</template>

<style scoped lang="scss">
.i-timeline-tag {
  display: inline-flex;
  align-items: center;
  font-size: 14px;
  line-height: 22px;
  color: $font-deep-color;

  &.is-current {
    font-size: 16px;
    font-weight: 700;
    line-height: 24px;
  }
}
</style>
