<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import bizView from './biz.vue';
import globalView from './global.vue';

const router = useRouter();
const route = useRoute();

const viewActive = computed({
  get() {
    return (route.params.view as string) || 'global';
  },
  set(value) {
    router.push({ params: { view: value } });
  },
});

const viewComps: Record<string, any> = {
  biz: bizView,
  global: globalView,
};
</script>

<template>
  <component :is="viewComps[viewActive]" />
</template>

<style lang="scss" scoped></style>
