<script setup lang="ts">
import { computed } from 'vue';

interface IProps {
  theme?: 'default' | 'primary' | 'success' | 'warning' | 'danger';
  href: string;
  disabled?: boolean;
  underline?: boolean;
  target?: '_self' | '_blank' | '_parent' | '_top';
  size?: 'small' | 'medium' | 'large';
}

defineOptions({ name: 'HcmLink' });
const props = withDefaults(defineProps<IProps>(), {
  theme: 'default',
  disabled: false,
  underline: false,
  target: '_self',
  size: 'medium',
});

const computedClass = computed(() => {
  return {
    [`is-${props.size}`]: props.size,
  };
});
</script>

<template>
  <bk-link v-bind="props" :class="computedClass">
    <slot></slot>
  </bk-link>
</template>

<style scoped lang="scss">
.is-small {
  font-size: 12px;
}
.is-medium {
  font-size: 14px;
}
.is-large {
  font-size: 16px;
}
</style>
