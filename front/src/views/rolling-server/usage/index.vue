<script setup lang="ts">
import { computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { IView } from './typings';

import orderView from './order.vue';

const router = useRouter();
const route = useRoute();

const viewActive = computed<IView>({
  get() {
    return (route.params.view as IView) || IView.ORDER;
  },
  set(value) {
    router.push({ params: { view: value } });
  },
});

const viewComps: Record<string, any> = {
  order: orderView,
};
</script>

<template>
  <component :is="viewComps[viewActive]" />
</template>

<style lang="scss" scoped></style>
