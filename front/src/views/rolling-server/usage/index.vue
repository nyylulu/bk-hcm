<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import appliedView from './applied.vue';

const router = useRouter();
const route = useRoute();

const viewActive = computed({
  get() {
    return (route.params.view as string) || 'applied';
  },
  set(value) {
    router.push({ params: { view: value } });
  },
});

const viewComps: Record<string, any> = {
  applied: appliedView,
};
</script>

<template>
  <component :is="viewComps[viewActive]" />
</template>

<style lang="scss" scoped></style>
