<script setup>
import { ref, nextTick, useTemplateRef, watch, onMounted } from 'vue';

const props = defineProps({
  text: {
    type: String,
    default: '',
  },
  maxWidth: {
    type: String,
    default: '100%',
  },
});

const containerRef = useTemplateRef('containerRef');
const textRef = useTemplateRef('textRef');
const isExpanded = ref(false);
const showExpandButton = ref(false);

// 检测文本是否被省略
const checkTextOverflow = () => {
  nextTick(() => {
    if (textRef.value && containerRef.value) {
      const textElement = textRef.value;
      // 检查文本是否溢出：scrollWidth > clientWidth
      const isOverflowing = textElement.scrollWidth > textElement.clientWidth;

      // 如果文本被省略且未展开，显示展开按钮
      showExpandButton.value = isOverflowing;
    }
  });
};

watch(
  () => props.text,
  () => {
    checkTextOverflow();
  },
);

// 切换展开/收起状态
const toggleExpand = () => {
  isExpanded.value = !isExpanded.value;
};

// 处理文本点击（可选，如果想点击文本也能切换）
const handleTextClick = () => {
  toggleExpand();
};

onMounted(() => {
  checkTextOverflow();
});
</script>

<template>
  <div class="text-container" :class="{ expanded: isExpanded }" ref="containerRef">
    <span ref="textRef" class="text-content" :class="{ expanded: isExpanded }" @click="handleTextClick">
      {{ text }}
    </span>
    <div v-if="showExpandButton" class="expand-button" @click="toggleExpand">
      {{ isExpanded ? '收起' : '展开' }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
.text-container {
  display: inline-flex;
  align-items: center;
  position: relative;
  width: 100%;

  &.expanded {
    flex-direction: column;
    align-items: flex-start;
  }
}

.text-content {
  max-width: 100%;
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.3s ease;

  &.expanded {
    overflow: visible;
    white-space: pre-wrap;
  }
}

.expand-button {
  margin-left: 11px;
  font-size: 14px;
  line-height: 22px;
  cursor: pointer;
  color: #3a84ff;
  width: 50px;
  flex-shrink: 0;
}
</style>
