<template>
  <section :class="!props.noShadow ? 'home' : undefined">
    <slot name="title">
      <div class="panel-title-container">
        <span class="title">{{ title }}</span>
        <slot name="title-extra"></slot>
      </div>
    </slot>
    <slot></slot>
  </section>
</template>

<script setup lang="ts">
interface IProps {
  title?: string;
  noShadow?: boolean;
}

const props = defineProps<IProps>();
</script>

<style lang="scss" scoped>
.home {
  width: 100%;
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  padding: 16px 24px;
}

.title {
  font-size: 14px;
  color: #313238;
  font-weight: 700;
  line-height: 22px;
  display: block;
}

.panel-title-container {
  display: flex;
  align-items: baseline;
  margin-bottom: 16px;
}
</style>
