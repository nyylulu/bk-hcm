<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import { Spinner } from 'bkui-vue/lib/icon';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';

import { isEqual } from 'lodash';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import http from '@/http';
import type { IQueryResData } from '@/typings';

interface IProps {
  params: {
    device_type: string;
    require_type: number;
    region: string;
    zone: string;
    vpc?: string;
    subnet?: string;
    charge_type?: string;
  };
}
type MaxResourceCapacity = IProps['params'];

interface ICvmCapacityItem {
  region: string;
  zone: string;
  vpc: string;
  subnet: string;
  max_num: number;
  max_info: { key: string; value: number }[];
}

const props = defineProps<IProps>();

const { t } = useI18n();

const loading = ref(false);
const list = ref<ICvmCapacityItem[]>([]);
const getList = async (params: MaxResourceCapacity) => {
  const { region, zone, device_type, require_type } = params;
  if (!(region && zone && device_type && require_type)) return;

  loading.value = true;
  try {
    const res: IQueryResData<{ info: ICvmCapacityItem[] }> = await http.post(
      '/api/v1/woa/config/find/cvm/capacity',
      params,
    );
    const { info = [] } = res.data;
    list.value = info.map((item) => ({ ...item, zoneCn: getZoneCn(item.zone) }));
  } catch (error) {
    console.error(error);
    list.value = [];
  } finally {
    loading.value = false;
  }
};

watch(
  () => props.params,
  (newVal, oldVal) => {
    if (!isEqual(newVal, oldVal)) {
      list.value = [];
      getList(newVal);
    }
  },
  { deep: true, immediate: true },
);
</script>

<template>
  <div class="cvm-max-capacity-wrapper">
    <!-- empty-state -->
    <div v-if="!list.length" class="flex-row align-items-center">
      <span>{{ t('最大可申请量：') }}</span>
      <span class="number error">0</span>
      <spinner v-if="loading" class="ml4" />
    </div>
    <!-- show-state -->
    <grid-container v-else fixed :column="4" gap="0">
      <template v-for="({ zone, max_num: maxNum, max_info: maxInfo }, index) in list" :key="index">
        <span>{{ getZoneCn(zone) }}</span>
        <span>{{ t('最大可申请量：') }}</span>
        <span class="number" :class="{ error: maxNum === 0 }">{{ maxNum }}</span>

        <bk-popover theme="light" ext-cls="detail-popover">
          <span class="detail-enter">{{ t('（计算明细）') }}</span>
          <template #content>
            <grid-container :column="2" :gap="[0, 8]">
              <template v-for="{ key, value } in maxInfo" :key="key">
                <span>{{ key }}</span>
                <span class="number" :class="{ error: value === 0 }">{{ value }}</span>
              </template>
            </grid-container>
          </template>
        </bk-popover>
      </template>
    </grid-container>
  </div>
</template>

<style scoped lang="scss">
.cvm-max-capacity-wrapper {
  font-size: 12px;
  color: #909399;
  line-height: 20px;

  .detail-enter {
    cursor: pointer;
  }
}

.number {
  font-size: 12px;
  color: $primary-color;

  &.error {
    color: $danger-color;
  }
}
</style>
