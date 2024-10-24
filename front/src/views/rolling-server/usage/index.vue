<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { IView } from './typings';

import appliedView from './applied.vue';

const router = useRouter();
const route = useRoute();

const viewActive = computed<IView>({
  get() {
    return (route.params.view as IView) || IView.APPLIED;
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
