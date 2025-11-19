<script setup lang="ts">
import { Plus } from 'bkui-vue/lib/icon';

defineProps<{
  disabled?: boolean;
}>();

const emit = defineEmits<{
  click: [];
}>();

const handleClick = () => {
  emit('click');
};
</script>
<template>
  <bk-button class="create-button" type="primary" :disabled="disabled" @click="handleClick">
    <slot>
      <plus class="icon" />
      添加配置
    </slot>
  </bk-button>
</template>
<style scoped lang="scss">
.create-button {
  .icon {
    font-size: 20px;
  }
}
</style>
