<script setup lang="ts">
import AlertIcon from '@/assets/image/alert.svg';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusFailure from '@/assets/image/failed-account.png';
import ResultDefault from '@/assets/image/result-default.svg';
import { Spinner } from 'bkui-vue/lib/icon';
import { computed, ref } from 'vue';

const props = defineProps<{
  errorMessage?: string;
  type: 'success' | 'default' | 'failed' | 'loading';
  text: string;
}>();

const TYPE_ICON = ref<Record<string, any>>({
  success: StatusSuccess,
  default: ResultDefault,
  failed: StatusFailure,
});

const renderIcon = computed(() => TYPE_ICON.value[props.type]);
</script>

<template>
  <div class="status-text">
    <spinner v-if="type === 'loading'" fill="#3A84FF" width="14" height="14" />
    <img v-else :src="renderIcon" width="14" height="14" alt="status-icon" class="mr6" />
    <span>{{ text }}</span>
    <img v-if="errorMessage" class="failed-tips" :src="AlertIcon" v-bk-tooltips="{ content: errorMessage }" />
  </div>
</template>

<style lang="scss" scoped>
.status-text {
  display: flex;
  align-items: center;

  .failed-tips {
    cursor: pointer;
    width: 12px;
    height: 12px;
    margin-left: 3px;
    margin-top: 1px;
  }
}
</style>
