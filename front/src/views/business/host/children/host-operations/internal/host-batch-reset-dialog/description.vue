<script setup lang="ts">
import { computed, h, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import { AngleUp, AngleDown } from 'bkui-vue/lib/icon';
import HcmLink from '@/components/hcm-link/index.vue';

const { t } = useI18n();

const isCollapse = ref(false);
const renderIcon = computed(() =>
  isCollapse.value ? h(AngleDown, { width: 20, height: 20 }) : h(AngleUp, { width: 20, height: 20 }),
);
const toggle = () => {
  isCollapse.value = !isCollapse.value;
};
</script>

<template>
  <div class="i-description">
    <strong>{{ t('重要说明') }}</strong>
    <div class="content" :class="{ 'is-collapse': isCollapse }">
      <p>1. {{ t('重装的主机，建议在配置平台的空闲机模块中，单次重装仅限500个IP') }}</p>
      <p>
        2. {{ t('仅支持云主机的重装，IDC物理机（自研云）重装请到') }}
        <hcm-link theme="primary" size="small" target="_blank" href="https://server.woa.com">
          https://server.woa.com
        </hcm-link>
        {{ t('操作') }}
      </p>
      <p>3. {{ t('主机的主负责人或备份负责人，必须为当前的执行人') }}</p>
      <p>4. {{ t('建议对数据做好相关备份后重装') }}</p>
      <p>5. {{ t('重装后，主机系统盘内的所有数据将被清除，恢复到初始状态，该操作不可恢复，请谨慎操作') }}</p>
      <p>
        6. {{ t('主机数据盘的数据仍保留，但重装系统后需要手动挂载才能使用') }}
        <hcm-link
          theme="primary"
          size="small"
          target="_blank"
          href="https://cloud.tencent.com/document/product/213/17487"
        >
          {{ t('参考文档') }}
        </hcm-link>
      </p>
      <p>7. {{ t('重装后，需要人工初始化，如安装 gse_agent，bf 等') }}</p>
    </div>
    <div class="i-op-wrap">
      <bk-button theme="primary" text @click="toggle">
        <component :is="renderIcon" />
        <span>{{ isCollapse ? t('展开全部') : t('收起') }}</span>
      </bk-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.i-description {
  padding: 8px 16px;
  line-height: 20px;
  background: #f5f7fa;
  border-radius: 2px;
  font-size: 12px;

  strong {
    display: inline-block;
    margin-bottom: 4px;
    font-weight: 700;
    color: #313238;
  }

  .content {
    overflow: hidden;
    max-height: 160px;
    transition: all 0.3s ease-in-out;

    &.is-collapse {
      max-height: 40px;
    }
  }

  .i-op-wrap {
    width: 100%;
    text-align: center;
  }
}
</style>
