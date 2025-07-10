<script setup lang="ts">
import { ModelPropertyDisplay } from '@/model/typings';
import { VendorEnum } from '@/common/constant';
import { get } from 'lodash';

import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';

interface IProps {
  column?: number;
  fields: ModelPropertyDisplay[];
  details: any;
}

withDefaults(defineProps<IProps>(), {
  column: 2,
});

const getDisplayCompProps = (field: ModelPropertyDisplay) => {
  const { id } = field;
  if (id === 'spec.region') {
    return {
      vendor: VendorEnum.ZIYAN, // 自研云功能
    };
  }
  return {};
};
</script>

<template>
  <grid-container :column="column">
    <grid-item v-for="field in fields" :key="field.id" :label="field.name">
      <bk-loading v-if="!details" size="mini" mode="spin" theme="primary" loading></bk-loading>
      <template v-else>
        <component v-if="field.render" :is="() => field.render(details)" />
        <display-value
          v-else
          :value="get(details, field.id)"
          :property="field"
          :display="field?.meta?.display"
          v-bind="getDisplayCompProps(field)"
        />
      </template>
    </grid-item>
  </grid-container>
</template>
