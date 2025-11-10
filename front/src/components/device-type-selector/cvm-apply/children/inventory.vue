<script setup lang="ts">
import { ref, computed, inject } from 'vue';
import { RequirementType } from '@/store/config/requirement';
import { useCvmDeviceStore, IManyCvmCapacityItem } from '@/store/cvm/device';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';

const props = defineProps<{
  maxLimit: number;
  region: string;
  zone: string;
  zoneList: string[];
  chargeType: string;
  deviceType: string;
  isActive: boolean;
}>();

const emit = defineEmits<{
  active: [deviceType: string];
}>();

const requireType = inject<RequirementType>('requireType');

const cvmDeviceStore = useCvmDeviceStore();

const loading = ref(false);

const capacityList = ref<IManyCvmCapacityItem[]>([]);

const maxNum = computed(() => Math.max(...capacityList.value?.map((item) => item.max_num)) ?? 0);

const currentZone = computed(() => {
  if (props.zone === 'all') {
    const capacity = capacityList.value.find((item) => item.max_num === maxNum.value);
    return capacity?.zone;
  }
  return props.zone;
});

const capacityState = computed(() => {
  const max = maxNum.value;
  const state = { class: '', text: '' };
  if (max > 50) {
    state.class = 'success';
    state.text = '库存充足';
  } else if (max >= 10 && max <= 50) {
    state.class = 'danger';
    state.text = '库存紧张';
  } else {
    state.class = '';
    state.text = '库存不足';
  }
  return state;
});

const handleViewInventory = async () => {
  emit('active', props.deviceType);
  if (loading.value) {
    return;
  }

  loading.value = true;

  const params = {
    region: props.region,
    require_type: requireType,
    zones: props.zoneList,
    device_types: [props.deviceType],
    charge_type: props.chargeType,
  };

  try {
    const list = await cvmDeviceStore.getManyCvmCapacity(params);
    capacityList.value = list ?? [];
  } finally {
    loading.value = false;
  }
};

const afterHidden = () => {
  emit('active', null);
};
const afterShow = () => {
  emit('active', props.deviceType);
};
</script>

<template>
  <div :class="['max-limit', capacityState.class]">
    <span class="num">{{ maxLimit }}</span>
    <bk-popover
      theme="light"
      placement="right"
      trigger="click"
      ref="popover"
      @after-show="afterShow"
      @after-hidden="afterHidden"
      max-height="500px"
      max-width="300px"
    >
      <!-- 外层的div用于占位，当icon隐藏时popover的箭头保持正常 -->
      <div><i class="hcm-icon bkhcm-icon-file details-icon" @click="handleViewInventory"></i></div>
      <template #content>
        <bk-loading theme="primary" mode="spin" size="mini" v-if="loading" />
        <dl class="popover-content" v-else>
          <dt class="title">
            <span class="text">库存情况</span>
            <bk-tag size="small" radius="4px" :theme="capacityState.class">{{ capacityState.text }}</bk-tag>
          </dt>
          <dd
            :class="['zone-item', { current: currentZone === item.zone }]"
            v-for="item in capacityList"
            :key="item.zone"
          >
            <div class="zone-name">{{ getZoneCn(item.zone) }}</div>
            <div :class="['zone-num', { zero: item.max_num === 0 }]">{{ item.max_num }}</div>
          </dd>
        </dl>
      </template>
    </bk-popover>
  </div>
</template>

<style scoped lang="scss">
.max-limit {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;

  .details-icon {
    display: none;
    font-size: 14px;
    color: #3a84ff;
    cursor: pointer;
  }
}

.popover-content {
  .title {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 0 6px;
    margin-bottom: 4px;

    .text {
      font-weight: 700;
      font-size: 12px;
    }
  }

  .zone-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 24px;
    font-size: 12px;
    padding: 0 6px;
    border-radius: 2px;

    .zone-num {
      height: 16px;
      min-width: 24px;
      text-align: center;
      color: #4d4f56;
      font-family: Arial, sans-serif;
      background: #f0f1f5;
      border-radius: 2px;
      margin: 4px 0;

      &.zero {
        color: #c4c6cc;
        background: #f5f7fa;
      }
    }

    &.current {
      background: #f0f5ff;

      .zone-num {
        color: #fff;
        background: #a3c5fd;
      }
    }
  }
}
</style>
